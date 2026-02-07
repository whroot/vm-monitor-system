package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// VM 虚拟机模型
type VM struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name       string    `gorm:"type:varchar(200);not null"`
	IP         string    `gorm:"type:inet"`
	OSType     string    `gorm:"type:varchar(20)"`
	OSVersion  string    `gorm:"type:varchar(100)"`
	CPUCores   int
	MemoryGB   int
	DiskGB     int
	Status     string `gorm:"type:varchar(20);not null;default:'unknown'"`
	PowerState string `gorm:"type:varchar(20)"`
	HostID     string `gorm:"type:varchar(100)"`
	HostName   string `gorm:"type:varchar(200)"`
	IsDeleted  bool   `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (VM) TableName() string {
	return "vms"
}

// VMMetrics 实时指标模型
type VMMetrics struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VMID           uuid.UUID `gorm:"type:uuid;not null"`
	CPUUsage       float64   `gorm:"type:numeric(5,2);not null;default:0"`
	MemoryUsage    float64   `gorm:"type:numeric(5,2);not null;default:0"`
	DiskUsage      float64   `gorm:"type:numeric(5,2);not null;default:0"`
	DiskReadMbps   float64   `gorm:"type:numeric(10,2);default:0"`
	DiskWriteMbps  float64   `gorm:"type:numeric(10,2);default:0"`
	NetworkInMbps  float64   `gorm:"type:numeric(10,2);default:0"`
	NetworkOutMbps float64   `gorm:"type:numeric(10,2);default:0"`
	Temperature    float64   `gorm:"type:numeric(5,2);default:0"`
	PowerState     string    `gorm:"type:varchar(20);default:'poweredOn'"`
	RecordedAt     time.Time
	CreatedAt      time.Time
}

func (VMMetrics) TableName() string {
	return "vm_metrics"
}

// VMMetricsHistory 历史指标模型
type VMMetricsHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VMID           uuid.UUID `gorm:"type:uuid;not null"`
	CPUUsage       float64   `gorm:"type:numeric(5,2);not null;default:0"`
	MemoryUsage    float64   `gorm:"type:numeric(5,2);not null;default:0"`
	DiskUsage      float64   `gorm:"type:numeric(5,2);not null;default:0"`
	DiskReadMbps   float64   `gorm:"type:numeric(10,2);default:0"`
	DiskWriteMbps  float64   `gorm:"type:numeric(10,2);default:0"`
	NetworkInMbps  float64   `gorm:"type:numeric(10,2);default:0"`
	NetworkOutMbps float64   `gorm:"type:numeric(10,2);default:0"`
	Temperature    float64   `gorm:"type:numeric(5,2);default:0"`
	RecordedAt     time.Time
}

func (VMMetricsHistory) TableName() string {
	return "vm_metrics_history"
}

func main() {
	// 连接数据库
	dsn := "host=localhost user=postgres password=postgres dbname=vm_monitoring port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// VM配置
	vmConfigs := []struct {
		name      string
		ip        string
		os        string
		osVersion string
		cpuCores  int
		memoryGB  int
		diskGB    int
		hostName  string
	}{
		{"web-server-01", "192.168.1.101", "Linux", "Ubuntu 22.04 LTS", 4, 8, 100, "esxi-host-01"},
		{"web-server-02", "192.168.1.102", "Linux", "Ubuntu 22.04 LTS", 4, 8, 100, "esxi-host-01"},
		{"app-server-01", "192.168.1.103", "Linux", "CentOS 8", 8, 16, 200, "esxi-host-02"},
		{"app-server-02", "192.168.1.104", "Linux", "CentOS 8", 8, 16, 200, "esxi-host-02"},
		{"db-server-01", "192.168.1.105", "Linux", "Ubuntu 20.04 LTS", 16, 32, 500, "esxi-host-03"},
		{"db-server-02", "192.168.1.106", "Linux", "Ubuntu 20.04 LTS", 16, 32, 500, "esxi-host-03"},
		{"cache-server-01", "192.168.1.107", "Linux", "Redis 7.0", 4, 16, 50, "esxi-host-01"},
		{"win-server-01", "192.168.1.108", "Windows", "Windows Server 2022", 8, 16, 300, "esxi-host-02"},
		{"win-server-02", "192.168.1.109", "Windows", "Windows Server 2019", 4, 8, 200, "esxi-host-02"},
		{"monitoring-01", "192.168.1.110", "Linux", "Ubuntu 22.04 LTS", 4, 8, 150, "esxi-host-01"},
	}

	fmt.Println("开始创建10台VM...")

	// 创建VM并生成数据
	for i, config := range vmConfigs {
		// 创建VM
		vm := VM{
			ID:         uuid.New(),
			Name:       config.name,
			IP:         config.ip,
			OSType:     config.os,
			OSVersion:  config.osVersion,
			CPUCores:   config.cpuCores,
			MemoryGB:   config.memoryGB,
			DiskGB:     config.diskGB,
			Status:     "online",
			PowerState: "poweredOn",
			HostID:     fmt.Sprintf("host-%d", i%3+1),
			HostName:   config.hostName,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := db.Create(&vm).Error; err != nil {
			fmt.Printf("创建VM %s 失败: %v\n", config.name, err)
			continue
		}

		fmt.Printf("✓ 创建VM: %s (ID: %s)\n", vm.Name, vm.ID)

		// 生成实时指标数据
		metrics := VMMetrics{
			ID:             uuid.New(),
			VMID:           vm.ID,
			CPUUsage:       round(rand.Float64()*80 + 10),  // 10-90%
			MemoryUsage:    round(rand.Float64()*70 + 20),  // 20-90%
			DiskUsage:      round(rand.Float64()*60 + 30),  // 30-90%
			DiskReadMbps:   round(rand.Float64()*100 + 10), // 10-110 MB/s
			DiskWriteMbps:  round(rand.Float64()*80 + 5),   // 5-85 MB/s
			NetworkInMbps:  round(rand.Float64()*50 + 5),   // 5-55 Mbps
			NetworkOutMbps: round(rand.Float64()*40 + 5),   // 5-45 Mbps
			Temperature:    round(rand.Float64()*30 + 35),  // 35-65°C
			PowerState:     "poweredOn",
			RecordedAt:     time.Now(),
			CreatedAt:      time.Now(),
		}

		if err := db.Create(&metrics).Error; err != nil {
			fmt.Printf("  创建实时指标失败: %v\n", err)
		} else {
			fmt.Printf("  ✓ 实时指标: CPU %.1f%%, 内存 %.1f%%\n", metrics.CPUUsage, metrics.MemoryUsage)
		}

		// 生成10条历史数据（过去24小时，每小时一条）
		fmt.Printf("  生成10条历史数据...\n")
		for j := 0; j < 10; j++ {
			history := VMMetricsHistory{
				ID:             uuid.New(),
				VMID:           vm.ID,
				CPUUsage:       round(rand.Float64()*80 + 10),
				MemoryUsage:    round(rand.Float64()*70 + 20),
				DiskUsage:      round(rand.Float64()*60 + 30),
				DiskReadMbps:   round(rand.Float64()*100 + 10),
				DiskWriteMbps:  round(rand.Float64()*80 + 5),
				NetworkInMbps:  round(rand.Float64()*50 + 5),
				NetworkOutMbps: round(rand.Float64()*40 + 5),
				Temperature:    round(rand.Float64()*30 + 35),
				RecordedAt:     time.Now().Add(-time.Duration(9-j) * time.Hour), // 过去9小时到现在，每小时一条
			}

			if err := db.Create(&history).Error; err != nil {
				fmt.Printf("    历史数据 %d 失败: %v\n", j+1, err)
			}
		}
		fmt.Printf("  ✓ 已生成10条历史数据\n\n")
	}

	fmt.Println("✅ 数据生成完成！")
	fmt.Println("\n统计:")

	var vmCount int64
	db.Model(&VM{}).Where("is_deleted = false").Count(&vmCount)
	fmt.Printf("- VM总数: %d\n", vmCount)

	var metricsCount int64
	db.Model(&VMMetrics{}).Count(&metricsCount)
	fmt.Printf("- 实时指标: %d\n", metricsCount)

	var historyCount int64
	db.Model(&VMMetricsHistory{}).Count(&historyCount)
	fmt.Printf("- 历史数据: %d\n", historyCount)
}

// round 保留一位小数
func round(val float64) float64 {
	return float64(int(val*10)) / 10
}
