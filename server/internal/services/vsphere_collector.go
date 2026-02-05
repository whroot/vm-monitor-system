package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"vm-monitoring-system/internal/logger"
	"vm-monitoring-system/internal/models"
)

// VSphereCollector vSphere数据采集器
type VSphereCollector struct {
	db           *gorm.DB
	client       *govmomi.Client
	viewManager  *view.Manager
	containerView *view.ContainerView
	config       *VSphereConfig
	isRunning    bool
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// VSphereConfig vSphere配置
type VSphereConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Insecure    bool   `json:"insecure"`
	CollectInterval time.Duration `json:"collectInterval"`
	BatchSize   int    `json:"batchSize"`
}

// MetricValue 指标值
type MetricValue struct {
	VMID      string
	Timestamp time.Time
	Metric    string
	Value     float64
}

// NewVSphereCollector 创建vSphere采集器
func NewVSphereCollector(db *gorm.DB, config *VSphereConfig) *VSphereCollector {
	if config.CollectInterval == 0 {
		config.CollectInterval = 30 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}

	return &VSphereCollector{
		db:        db,
		config:    config,
		stopChan:  make(chan struct{}),
		isRunning: false,
	}
}

// Start 启动采集器
func (c *VSphereCollector) Start() error {
	if c.isRunning {
		return fmt.Errorf("采集器已在运行")
	}

	// 连接vCenter
	if err := c.connect(); err != nil {
		return fmt.Errorf("连接vCenter失败: %w", err)
	}
	defer c.disconnect()

	// 创建容器视图
	if err := c.createContainerView(); err != nil {
		return fmt.Errorf("创建容器视图失败: %w", err)
	}
	defer c.containerView.Destroy(context.Background())

	c.isRunning = true
	log.Println("vSphere采集器已启动，采集间隔:", c.config.CollectInterval)

	// 启动采集循环
	c.collectLoop()

	return nil
}

// Stop 停止采集器
func (c *VSphereCollector) Stop() {
	if !c.isRunning {
		return
	}

	log.Println("正在停止vSphere采集器...")
	close(c.stopChan)
	c.wg.Wait()
	c.isRunning = false
	log.Println("vSphere采集器已停止")
}

// connect 连接vCenter
func (c *VSphereCollector) connect() error {
	url := fmt.Sprintf("https://%s:%d/sdk", c.config.Host, c.config.Port)
	ctx := context.Background()

	// 创建客户端
	client, err := govmomi.NewClient(ctx, url, c.config.Insecure)
	if err != nil {
		return fmt.Errorf("创建客户端失败: %w", err)
	}

	// 登录
	err = client.Login(ctx, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	c.client = client
	log.Printf("已连接到vCenter: %s", c.config.Host)

	return nil
}

// disconnect 断开连接
func (c *VSphereCollector) disconnect() {
	if c.client != nil {
		ctx := context.Background()
		c.client.Logout(ctx)
		c.client = nil
	}
}

// createContainerView 创建容器视图
func (c *VSphereCollector) createContainerView() error {
	ctx := context.Background()
	c.viewManager = view.NewManager(c.client.Client)

	// 创建根容器视图
	var err error
	c.containerView, err = c.viewManager.CreateContainerView(ctx, c.client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return fmt.Errorf("创建容器视图失败: %w", err)
	}

	return nil
}

// collectLoop 采集循环
func (c *VSphereCollector) collectLoop() {
	ticker := time.NewTicker(c.config.CollectInterval)
	defer ticker.Stop()

	// 立即执行一次采集
	c.collectMetrics()

	for {
		select {
		case <-ticker.C:
			c.collectMetrics()
		case <-c.stopChan:
			return
		}
	}
}

// collectMetrics 采集指标
func (c *VSphereCollector) collectMetrics() {
	ctx := context.Background()
	log.Println("开始采集VM指标...")

	// 获取所有VM
	var vms []mo.VirtualMachine
	err := c.containerView.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name", "runtime.powerState", "summary.quickStats"}, &vms)
	if err != nil {
		log.Printf("获取VM列表失败: %v", err)
		return
	}

	if len(vms) == 0 {
		log.Println("未找到VM")
		return
	}

	log.Printf("找到 %d 个VM", len(vms))

	// 获取性能管理器
	perfManager := performance.NewManager(c.client.Client)

	// 分批采集
	for i := 0; i < len(vms); i += c.config.BatchSize {
		end := i + c.config.BatchSize
		if end > len(vms) {
			end = len(vms)
		}

		batch := vms[i:end]
		c.collectBatch(ctx, perfManager, batch)
	}

	log.Println("VM指标采集完成")
}

// collectBatch 批量采集
func (c *VSphereCollector) collectBatch(ctx context.Context, perfManager *performance.Manager, vms []mo.VirtualMachine) {
	// 构建计数器信息
	counterInfo := c.buildCounterInfo()

	// 转换为ManagedObjectReference
	refs := make([]types.ManagedObjectReference, len(vms))
	for i, vm := range vms {
		refs[i] = vm.Self
	}

	// 定义要采集的指标
	metricSpecs := []performance.MetricId{
		{CounterId: counterInfo["cpu.usage.average"], Instance: "*"},
		{CounterId: counterInfo["mem.usage.average"], Instance: "*"},
		{CounterId: counterInfo["disk.usage.average"], Instance: "*"},
		{CounterId: counterInfo["net.usage.average"], Instance: "*"},
		{CounterId: counterInfo["disk.read.average"], Instance: "*"},
		{CounterId: counterInfo["disk.write.average"], Instance: "*"},
	}

	// 查询性能数据
	query := new(performance.QuerySpec)
	query.Entity = refs
	query.MetricId = metricSpecs
	query.Start = time.Now().Add(-1 * time.Minute)
	query.End = time.Now()

	metricSeries, err := perfManager.Query(ctx, query)
	if err != nil {
		log.Printf("查询性能数据失败: %v", err)
		return
	}

	// 处理采集到的数据
	metrics := c.processMetricData(vms, metricSeries, counterInfo)

	// 保存到数据库
	if len(metrics) > 0 {
		c.saveMetrics(metrics)
	}
}

// buildCounterInfo 构建计数器信息
func (c *VSphereCollector) buildCounterInfo() map[string]int32 {
	ctx := context.Background()
	perfManager := performance.NewManager(c.client.Client)

	// 获取计数器信息
	counters, err := perfManager.CounterInfoByKey(ctx)
	if err != nil {
		log.Printf("获取计数器信息失败: %v", err)
		return nil
	}

	counterMap := make(map[string]int32)
	for key, counter := range counters {
		counterMap[counter.GroupInfo.Key+"."+counter.NameInfo.RollupType] = key
	}

	return counterMap
}

// processMetricData 处理指标数据
func (c *VSphereCollector) processMetricData(vms []mo.VirtualMachine, metricSeries []performance.MetricSeries, counterInfo map[string]int32) []MetricValue {
	metrics := []MetricValue{}

	for _, vm := range vms {
		vmID := vm.Self.Value

		// 更新VM基本信息
		c.updateVMInfo(vm)

		// 查找该VM的性能数据
		for _, series := range metricSeries {
			if series.Entity == vm.Self {
				// 处理每个指标值
				for _, value := range series.Value {
					if value == nil {
						continue
					}

					// 获取指标名称
					metricName := c.getMetricName(value.Id, counterInfo)
					if metricName == "" {
						continue
					}

					// 提取数值
					var metricValue float64
					switch v := value.(type) {
					case *performance.MetricSeriesSample:
						if len(v.Sample) > 0 {
							// 使用最新值
							metricValue = float64(v.Sample[len(v.Sample)-1])
						}
					}

					// 跳过无效值
					if metricValue <= 0 {
						continue
					}

					metrics = append(metrics, MetricValue{
						VMID:      vmID,
						Timestamp: time.Now(),
						Metric:    metricName,
						Value:     metricValue,
					})
				}
			}
		}
	}

	return metrics
}

// getMetricName 获取指标名称
func (c *VSphereCollector) getMetricName(counterID int32, counterInfo map[string]int32) string {
	// 根据counterID查找对应的指标名称
	counter := counterInfo[counterID]
	
	// 简化指标名称
	switch counter {
	case counterInfo["cpu.usage.average"]:
		return "cpu_usage"
	case counterInfo["mem.usage.average"]:
		return "memory_usage"
	case counterInfo["disk.usage.average"]:
		return "disk_usage"
	case counterInfo["net.usage.average"]:
		return "network_usage"
	case counterInfo["disk.read.average"]:
		return "disk_io_read"
	case counterInfo["disk.write.average"]:
		return "disk_io_write"
	default:
		return ""
	}
}

// updateVMInfo 更新VM信息
func (c *VSphereCollector) updateVMInfo(vm mo.VirtualMachine) {
	// 查找数据库中是否已存在该VM
	var dbVM models.VM
	err := c.db.Where("vmware_id = ?", vm.Self.Value).First(&dbVM).Error

	powerState := string(vm.Runtime.PowerState)
	status := "unknown"
	if powerState == "poweredOn" {
		status = "online"
	} else if powerState == "poweredOff" {
		status = "offline"
	}

	if err == gorm.ErrRecordNotFound {
		// 创建新VM记录
		newVM := models.VM{
			VMwareID:     &vm.Self.Value,
			Name:         vm.Name,
			PowerState:   &powerState,
			Status:       status,
			LastSeen:     &[]time.Time{time.Now()}[0],
			IsDeleted:    false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if vm.Summary != nil && vm.Summary.QuickStats != nil {
			c.updateVMQuickStats(&newVM, vm.Summary.QuickStats)
		}

		if err := c.db.Create(&newVM).Error; err != nil {
			log.Printf("创建VM记录失败 (%s): %v", vm.Name, err)
		}
	} else if err == nil {
		// 更新现有VM记录
		updates := map[string]interface{}{
			"power_state": powerState,
			"status":      status,
			"last_seen":   time.Now(),
			"updated_at":  time.Now(),
		}

		if vm.Summary != nil && vm.Summary.QuickStats != nil {
			c.updateVMQuickStats(&dbVM, vm.Summary.QuickStats)
		}

		if err := c.db.Model(&dbVM).Updates(updates).Error; err != nil {
			log.Printf("更新VM记录失败 (%s): %v", vm.Name, err)
		}
	}
}

// updateVMQuickStats 更新VM快速统计信息
func (c *VSphereCollector) updateVMQuickStats(vm *models.VM, stats *object.VirtualMachineQuickStats) {
	// 更新CPU
	if stats.NumCpu != nil && *stats.NumCpu > 0 {
		vm.CPUCores = stats.NumCpu
	}

	// 更新内存
	if stats.HostMemoryUsage != nil && *stats.HostMemoryUsage > 0 {
		memoryGB := *stats.HostMemoryUsage / 1024
		vm.MemoryGB = &memoryGB
	}
}

// saveMetrics 保存指标到数据库
func (c *VSphereCollector) saveMetrics(metrics []MetricValue) {
	if len(metrics) == 0 {
		return
	}

	// 转换为时序数据结构
	metricData := make([]services.MetricData, len(metrics))
	for i, m := range metrics {
		metricData[i] = services.MetricData{
			ID:        uuid.New(),
			VMID:      m.VMID,
			Metric:    m.Metric,
			Value:     m.Value,
			Timestamp: m.Timestamp,
		}
	}

	// 保存到数据库
	err := c.db.SaveMetrics(metricData)
	if err != nil {
		log.Printf("保存指标数据失败: %v", err)
		return
	}

	log.Printf("已保存 %d 条指标数据", len(metrics))
}

// GetVMStatus 获取VM状态
func (c *VSphereCollector) GetVMStatus(vmID string) (*VMStatus, error) {
	ctx := context.Background()
	
	// 查找VM对象
	vmRef := types.ManagedObjectReference{
		Type:  "VirtualMachine",
		Value: vmID,
	}

	vm := object.NewVirtualMachine(c.client.Client, vmRef)

	// 获取VM状态
	summary, err := vm.Summary(ctx)
	if err != nil {
		return nil, err
	}

	status := &VMStatus{
		VMID:      vmID,
		Name:      summary.Config.Name,
		PowerState: string(summary.Runtime.PowerState),
		IPAddress: summary.Guest.IpAddress,
		UpdatedAt: time.Now(),
	}

	if summary.QuickStats != nil {
		status.CPUUsage = summary.QuickStats.OverallCpuUsage
		status.MemoryUsage = summary.QuickStats.GuestMemoryUsage
		status.DiskUsage = summary.QuickStats.OverallCpuUsage
		status.Uptime = time.Duration(summary.QuickStats.UptimeSeconds) * time.Second
	}

	return status, nil
}

// VMStatus VM状态
type VMStatus struct {
	VMID        string
	Name        string
	PowerState  string
	IPAddress   string
	CPUUsage    *int32
	MemoryUsage *int32
	DiskUsage   *int32
	Uptime      time.Duration
	UpdatedAt   time.Time
}
