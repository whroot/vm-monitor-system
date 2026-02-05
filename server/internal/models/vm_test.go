package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVMTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&VM{}, &VMGroup{}, &VMGroupMember{})

	return db, func() {
		db.Close()
	}
}

func TestVM_TableName(t *testing.T) {
	vm := VM{}
	assert.Equal(t, "vms", vm.TableName())
}

func TestVMGroup_TableName(t *testing.T) {
	group := VMGroup{}
	assert.Equal(t, "vm_groups", group.TableName())
}

func TestVMGroupMember_TableName(t *testing.T) {
	member := VMGroupMember{}
	assert.Equal(t, "vm_group_members", member.TableName())
}

func TestVM_CRUD(t *testing.T) {
	db, teardown := setupVMTestDB()
	defer teardown()

	t.Run("CreateVM", func(t *testing.T) {
		vm := VM{
			ID:       uuid.New(),
			Name:     "test-vm",
			IP:       StringPtr("192.168.1.100"),
			Status:   "running",
			OSType:   StringPtr("Linux"),
		}

		err := db.Create(&vm).Error

		assert.NoError(t, err)
		assert.NotEmpty(t, vm.ID)
	})

	t.Run("ReadVM", func(t *testing.T) {
		vm := VM{
			ID:     uuid.New(),
			Name:   "test-vm-read",
			Status: "running",
		}
		db.Create(&vm)

		var found VM
		err := db.First(&found, vm.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, "test-vm-read", found.Name)
	})

	t.Run("UpdateVM", func(t *testing.T) {
		vm := VM{
			ID:     uuid.New(),
			Name:   "test-vm-update",
			Status: "running",
		}
		db.Create(&vm)

		err := db.Model(&vm).Update("status", "stopped").Error

		assert.NoError(t, err)

		var found VM
		db.First(&found, vm.ID)
		assert.Equal(t, "stopped", found.Status)
	})

	t.Run("SoftDeleteVM", func(t *testing.T) {
		vm := VM{
			ID:     uuid.New(),
			Name:   "test-vm-delete",
			Status: "running",
		}
		db.Create(&vm)

		err := db.Model(&vm).Update("is_deleted", true).Error

		assert.NoError(t, err)

		var found VM
		err = db.Where("id = ?", vm.ID).Unscoped().First(&found).Error
		assert.NoError(t, err)
		assert.True(t, found.IsDeleted)
	})

	t.Run("FindByStatus", func(t *testing.T) {
		vms := []VM{
			{ID: uuid.New(), Name: "vm1", Status: "running"},
			{ID: uuid.New(), Name: "vm2", Status: "running"},
			{ID: uuid.New(), Name: "vm3", Status: "stopped"},
		}
		db.Create(&vms)

		var running []VM
		err := db.Where("status = ?", "running").Find(&running).Error

		assert.NoError(t, err)
		assert.Len(t, running, 2)
	})
}

func TestVMGroup_CRUD(t *testing.T) {
	db, teardown := setupVMTestDB()
	defer teardown()

	t.Run("CreateGroup", func(t *testing.T) {
		group := VMGroup{
			ID:   uuid.New(),
			Name: "Test Group",
			Type: "custom",
		}

		err := db.Create(&group).Error

		assert.NoError(t, err)
		assert.NotEmpty(t, group.ID)
	})

	t.Run("CreateSubGroup", func(t *testing.T) {
		parent := VMGroup{
			ID:   uuid.New(),
			Name: "Parent Group",
			Type: "folder",
		}
		db.Create(&parent)

		child := VMGroup{
			ID:       uuid.New(),
			Name:     "Child Group",
			Type:     "custom",
			ParentID: &parent.ID,
		}
		err := db.Create(&child).Error

		assert.NoError(t, err)

		var found VMGroup
		db.Preload("Children").First(&found, parent.ID)
		assert.Len(t, found.Children, 1)
	})
}

func TestVMGroupMember_CRUD(t *testing.T) {
	db, teardown := setupVMTestDB()
	defer teardown()

	t.Run("AddVMToGroup", func(t *testing.T) {
		vm := VM{
			ID:     uuid.New(),
			Name:   "test-vm",
			Status: "running",
		}
		db.Create(&vm)

		group := VMGroup{
			ID:   uuid.New(),
			Name: "Test Group",
			Type: "custom",
		}
		db.Create(&group)

		member := VMGroupMember{
			ID:      uuid.New(),
			VMID:    vm.ID,
			GroupID: group.ID,
		}
		err := db.Create(&member).Error

		assert.NoError(t, err)
	})

	t.Run("GetVMsInGroup", func(t *testing.T) {
		vm1 := VM{ID: uuid.New(), Name: "vm1", Status: "running"}
		vm2 := VM{ID: uuid.New(), Name: "vm2", Status: "running"}
		db.Create(&vm1, &vm2)

		group := VMGroup{
			ID:   uuid.New(),
			Name: "Test Group",
			Type: "custom",
		}
		db.Create(&group)

		members := []VMGroupMember{
			{ID: uuid.New(), VMID: vm1.ID, GroupID: group.ID},
			{ID: uuid.New(), VMID: vm2.ID, GroupID: group.ID},
		}
		db.Create(&members)

		var vms []VM
		db.Joins("JOIN vm_group_members ON vms.id = vm_group_members.vm_id").
			Where("vm_group_members.group_id = ?", group.ID).
			Find(&vms)

		assert.Len(t, vms, 2)
	})
}

func TestVMListRequest(t *testing.T) {
	req := VMListRequest{
		Page:     1,
		PageSize: 20,
		Status:   "running",
		OS:       "Linux",
		Keyword:  "test",
	}

	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 20, req.PageSize)
	assert.Equal(t, "running", req.Status)
	assert.Equal(t, "Linux", req.OS)
	assert.Equal(t, "test", req.Keyword)
}

func TestVMSummary(t *testing.T) {
	summary := VMSummary{
		Total:   100,
		Online:  80,
		Offline: 15,
		Error:   5,
	}

	assert.Equal(t, 100, summary.Total)
	assert.Equal(t, 80, summary.Online)
	assert.Equal(t, 15, summary.Offline)
	assert.Equal(t, 5, summary.Error)
}

func TestVMStatistics(t *testing.T) {
	stats := VMStatistics{}

	stats.Overview.Total = 100
	stats.Overview.Online = 75
	stats.Overview.Offline = 20
	stats.Overview.Error = 3
	stats.Overview.Unknown = 2

	assert.Equal(t, 100, stats.Overview.Total)
	assert.Equal(t, 75, stats.Overview.Online)
	assert.Equal(t, 20, stats.Overview.Offline)
	assert.Equal(t, 3, stats.Overview.Error)
	assert.Equal(t, 2, stats.Overview.Unknown)
}

func TestVMSyncRequest(t *testing.T) {
	t.Run("FullSync", func(t *testing.T) {
		req := VMSyncRequest{
			Type:    "full",
			ClusterID: "cluster-123",
		}
		assert.Equal(t, "full", req.Type)
	})

	t.Run("IncrementalSync", func(t *testing.T) {
		req := VMSyncRequest{
			Type:         "incremental",
			DatacenterID: "dc-456",
		}
		assert.Equal(t, "incremental", req.Type)
	})
}

func TestVMSyncResponse(t *testing.T) {
	now := time.Now()
	response := VMSyncResponse{
		SyncID:     "sync-123",
		Status:     "completed",
		StartedAt:  now,
		CompletedAt: &now,
		Result: &SyncResult{
			TotalVMs: 100,
			Added:    5,
			Updated:  10,
			Removed:  2,
			Failed:   0,
		},
	}

	assert.Equal(t, "sync-123", response.SyncID)
	assert.Equal(t, "completed", response.Status)
	assert.NotNil(t, response.Result)
	assert.Equal(t, 100, response.Result.TotalVMs)
}

func TestSyncResult(t *testing.T) {
	result := SyncResult{
		TotalVMs: 50,
		Added:    3,
		Updated:  5,
		Removed:  1,
		Failed:   0,
		Errors: []SyncError{
			{VMwareID: "vm-err-1", Error: "connection timeout"},
		},
	}

	assert.Equal(t, 50, result.TotalVMs)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "vm-err-1", result.Errors[0].VMwareID)
}

func TestVM_IPNetwork(t *testing.T) {
	db, teardown := setupVMTestDB()
	defer teardown()

	vm := VM{
		ID:     uuid.New(),
		Name:   "test-vm-ip",
		IP:     StringPtr("192.168.1.100"),
		Status: "running",
	}
	db.Create(&vm)

	var found VM
	db.First(&found, vm.ID)

	assert.NotNil(t, found.IP)
	assert.Equal(t, "192.168.1.100", *found.IP)
}

func TestVM_Pagination(t *testing.T) {
	db, teardown := setupVMTestDB()
	defer teardown()

	for i := 0; i < 50; i++ {
		db.Create(&VM{
			ID:     uuid.New(),
			Name:   "test-vm-" + string(rune(i)),
			Status: "running",
		})
	}

	var vms []VM
	err := db.Limit(10).Offset(0).Find(&vms).Error

	assert.NoError(t, err)
	assert.Len(t, vms, 10)
}
