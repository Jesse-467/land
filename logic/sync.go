package logic

import (
	"land/dao/redis"
	"time"

	"go.uber.org/zap"
)

// ViewCountSyncService 访问量同步服务
type ViewCountSyncService struct {
	syncInterval time.Duration // 同步间隔
	stopChan     chan bool     // 停止信号
}

// NewViewCountSyncService 创建访问量同步服务
// 参数:
//   - syncInterval: 同步间隔时间
//
// 返回值:
//   - *ViewCountSyncService: 同步服务实例
func NewViewCountSyncService(syncInterval time.Duration) *ViewCountSyncService {
	return &ViewCountSyncService{
		syncInterval: syncInterval,
		stopChan:     make(chan bool),
	}
}

// Start 启动同步服务
func (s *ViewCountSyncService) Start() {
	go s.syncLoop()
	zap.L().Info("ViewCountSyncService started",
		zap.Duration("sync_interval", s.syncInterval))
}

// Stop 停止同步服务
func (s *ViewCountSyncService) Stop() {
	close(s.stopChan)
	zap.L().Info("ViewCountSyncService stopped")
}

// syncLoop 同步循环
func (s *ViewCountSyncService) syncLoop() {
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performSync()
		case <-s.stopChan:
			return
		}
	}
}

// performSync 执行同步操作
func (s *ViewCountSyncService) performSync() {
	zap.L().Debug("Starting view count sync...")

	// 获取所有Redis中的访问量数据
	viewCounts, err := redis.GetAllPostViewCounts()
	if err != nil {
		zap.L().Error("Failed to get all post view counts", zap.Error(err))
		return
	}

	if len(viewCounts) == 0 {
		zap.L().Debug("No view counts to sync")
		return
	}

	// 同步到MySQL
	err = redis.SyncViewCountsToMySQL(viewCounts)
	if err != nil {
		zap.L().Error("Failed to sync view counts to MySQL", zap.Error(err))
		return
	}

	// 更新访问量有序集合
	err = redis.InitPostViewZSet(viewCounts)
	if err != nil {
		zap.L().Error("Failed to update post view ZSet", zap.Error(err))
		// 不返回错误，因为MySQL同步已经成功
	}

	zap.L().Info("View count sync completed",
		zap.Int("synced_posts", len(viewCounts)))
}

// ManualSync 手动同步（可用于API调用）
// 返回值:
//   - error: 可能的错误
func (s *ViewCountSyncService) ManualSync() error {
	zap.L().Info("Manual view count sync triggered")
	s.performSync()
	return nil
}
