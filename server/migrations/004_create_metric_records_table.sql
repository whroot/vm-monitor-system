-- 创建时序数据表（PostgreSQL with TimescaleDB扩展）
-- 注意：此脚本需要TimescaleDB扩展支持

-- 启用TimescaleDB扩展（如果未启用）
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 创建指标记录表
CREATE TABLE IF NOT EXISTS metric_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vm_id VARCHAR(100) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    tags JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 创建复合索引优化查询
CREATE INDEX IF NOT EXISTS idx_metric_records_vm_metric_time 
    ON metric_records (vm_id, metric, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_metric_records_metric_time 
    ON metric_records (metric, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_metric_records_time 
    ON metric_records (timestamp DESC);

-- 创建Hypertable（TimescaleDB特性）
-- 这将自动按时间分区，提高查询性能
SELECT create_hypertable('metric_records', 'timestamp',
    if_not_exists => TRUE,
    chunk_time_interval => interval '1 day'
);

-- 设置数据保留策略（可选）
-- 例如：保留最近30天的数据
-- SELECT add_retention_policy('metric_records', INTERVAL '30 days');

-- 创建压缩策略（可选）
-- 压缩7天前的数据以节省存储空间
-- SELECT add_compression_policy('metric_records', INTERVAL '7 days');

-- 创建聚合视图（可选）
-- 用于快速查询聚合数据
-- CREATE MATERIALIZED VIEW IF NOT EXISTS metric_hourly_summary
-- AS
-- SELECT
--     time_bucket('1 hour', timestamp) as bucket,
--     vm_id,
--     metric,
--     AVG(value) as avg_value,
--     MAX(value) as max_value,
--     MIN(value) as min_value,
--     COUNT(*) as count
-- FROM metric_records
-- GROUP BY bucket, vm_id, metric;

-- 为聚合视图创建索引
-- CREATE INDEX IF NOT EXISTS idx_metric_hourly_summary_bucket
--     ON metric_hourly_summary (bucket DESC, vm_id, metric);

-- 添加注释
COMMENT ON TABLE metric_records IS 'VM监控指标数据表（TimescaleDB Hypertable）';
COMMENT ON COLUMN metric_records.id IS '主键ID';
COMMENT ON COLUMN metric_records.vm_id IS '虚拟机ID';
COMMENT ON COLUMN metric_records.metric IS '指标名称 (cpu_usage, memory_usage, disk_usage, network_usage等)';
COMMENT ON COLUMN metric_records.value IS '指标值';
COMMENT ON COLUMN metric_records.timestamp IS '时间戳';
COMMENT ON COLUMN metric_records.tags IS '标签（JSON格式）';
COMMENT ON COLUMN metric_records.created_at IS '创建时间';

-- 创建指标类型枚举（可选，用于约束metric字段）
-- CREATE TYPE metric_type AS ENUM (
--     'cpu_usage',
--     'memory_usage',
--     'disk_usage',
--     'network_usage',
--     'disk_io_read',
--     'disk_io_write',
--     'network_rx',
--     'network_tx'
-- );

-- 如果使用枚举，可以添加约束
-- ALTER TABLE metric_records 
--     ADD CONSTRAINT chk_metric_type 
--     CHECK (metric::text = ANY(
--         SELECT enumlabel FROM pg_enum 
--         WHERE enumtypid = 'metric_type'::regtype
--     ));

-- 创建触发器用于数据清理（可选）
-- CREATE OR REPLACE FUNCTION cleanup_old_metrics()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     IF NEW.created_at < NOW() - INTERVAL '90 days' THEN
--         -- 删除90天前的数据
--         DELETE FROM metric_records 
--         WHERE timestamp < NEW.created_at - INTERVAL '90 days';
--     END IF;
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;

-- CREATE TRIGGER trigger_cleanup_old_metrics
--     AFTER INSERT ON metric_records
--     FOR EACH ROW
--     EXECUTE FUNCTION cleanup_old_metrics();
