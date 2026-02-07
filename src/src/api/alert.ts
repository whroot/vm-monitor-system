import apiClient from './client';
import { AlertRecord } from '../types/api';

const MOCK_MODE = false;

const mockAlerts: AlertRecord[] = [
  {
    id: 'alert_001',
    ruleId: 'rule_001',
    ruleName: 'CPU使用率过高',
    vmId: 'vm_001',
    vmName: 'prod-app-01',
    metric: 'cpu.usage.percent',
    severity: 'critical',
    triggerValue: 95.5,
    threshold: 80,
    triggeredAt: new Date(Date.now() - 1800000).toISOString(),
    status: 'active',
  },
  {
    id: 'alert_002',
    ruleId: 'rule_002',
    ruleName: '内存使用率告警',
    vmId: 'vm_002',
    vmName: 'prod-app-02',
    metric: 'memory.usage.percent',
    severity: 'high',
    triggerValue: 88.2,
    threshold: 85,
    triggeredAt: new Date(Date.now() - 3600000).toISOString(),
    status: 'acknowledged',
    acknowledgedByName: 'admin',
    acknowledgedAt: new Date(Date.now() - 3000000).toISOString(),
  },
];

export interface AlertRule {
  ID: string;
  Name: string;
  Description: string | null;
  Scope: string;
  Severity: string;
  Enabled: boolean;
  Cooldown: number;
  NotificationConfig: string;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface AlertStats {
  Total: number;
  Active: number;
  Critical: number;
  Warning: number;
}

export interface AlertListResponse {
  rules?: AlertRule[];
  records?: AlertRecord[];
  total?: number;
  page?: number;
  pageSize?: number;
}

export const alertApi = {
  listRules: async (params?: { page?: number; pageSize?: number }): Promise<AlertListResponse> => {
    if (MOCK_MODE) {
      return {
        rules: [
          { ID: 'rule_001', Name: 'CPU使用率告警', Description: 'CPU使用率超过80%', Scope: 'vm', Severity: 'warning', Enabled: true, Cooldown: 300, NotificationConfig: '{}', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString() },
          { ID: 'rule_002', Name: '内存使用率告警', Description: '内存使用率超过85%', Scope: 'vm', Severity: 'critical', Enabled: true, Cooldown: 300, NotificationConfig: '{}', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString() },
        ],
        total: 2,
        page: 1,
        pageSize: 10,
      };
    }
    const response = await apiClient.get('/alerts/rules', { params }) as { rules: AlertRule[]; total: number; page: number; pageSize: number };
    return {
      rules: response.rules,
      total: response.total,
      page: response.page,
      pageSize: response.pageSize,
    };
  },

  getRule: async (id: string): Promise<AlertRule> => {
    if (MOCK_MODE) {
      return { ID: 'rule_001', Name: 'CPU使用率告警', Description: 'CPU使用率超过80%', Scope: 'vm', Severity: 'warning', Enabled: true, Cooldown: 300, NotificationConfig: '{}', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString() };
    }
    const response = await apiClient.get(`/alerts/rules/${id}`) as AlertRule;
    return response;
  },

  createRule: async (data: { name: string; scope: string; severity: string }): Promise<{ ruleId: string }> => {
    if (MOCK_MODE) {
      return { ruleId: `rule_${Date.now()}` };
    }
    const response = await apiClient.post('/alerts/rules', data) as { ruleId: string };
    return response;
  },

  updateRule: async (id: string, data: Partial<AlertRule>): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.put(`/alerts/rules/${id}`, data);
  },

  deleteRule: async (id: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.delete(`/alerts/rules/${id}`);
  },

  listRecords: async (params?: { page?: number; pageSize?: number }): Promise<AlertListResponse> => {
    if (MOCK_MODE) {
      return {
        records: mockAlerts,
        total: mockAlerts.length,
        page: 1,
        pageSize: 10,
      };
    }
    const response = await apiClient.get('/alerts/records', { params }) as { records: AlertRecord[]; total: number; page: number; pageSize: number };
    return {
      records: response.records,
      total: response.total,
      page: response.page,
      pageSize: response.pageSize,
    };
  },

  getStats: async (): Promise<AlertStats> => {
    if (MOCK_MODE) {
      return { Total: 15, Active: 5, Critical: 2, Warning: 3 };
    }
    const response = await apiClient.get('/alerts/stats') as AlertStats;
    return response;
  },

  acknowledge: async (id: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.post(`/alerts/${id}/acknowledge`);
  },

  resolve: async (id: string, resolution: string): Promise<void> => {
    if (MOCK_MODE) return;
    await apiClient.post(`/alerts/${id}/resolve`, { resolution });
  },

  createTestAlert: async (data: { vmName: string; severity: string; metric: string; value: number }): Promise<{ alertId: string }> => {
    if (MOCK_MODE) {
      return { alertId: `alert_${Date.now()}` };
    }
    const response = await apiClient.post('/alerts/test', data) as { alertId: string };
    return response;
  },
};
