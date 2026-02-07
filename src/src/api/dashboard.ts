import client from './client';

export interface OverviewResponse {
  healthScore: number;
  healthTrend: string;
  lastUpdated: string;
  systemStatus: string;
  summary: {
    totalVMs: number;
    onlineVMs: number;
    offlineVMs: number;
    warningVMs: number;
    criticalVMs: number;
  };
  metrics: {
    cpu: {
      usagePercent: number;
      trend: string;
      trendValue: number;
    };
    memory: {
      usagePercent: number;
      trend: string;
      trendValue: number;
    };
    disk: {
      usagePercent: number;
      trend: string;
      trendValue: number;
    };
    network: {
      inboundMbps: number;
      outboundMbps: number;
      trend: string;
      trendValue: number;
    };
  };
  topResources?: {
    byCPU: Array<{ vmId: string; vmName: string; usagePercent: number }>;
    byMemory: Array<{ vmId: string; vmName: string; usagePercent: number }>;
  };
}

export interface VMStatusDistribution {
  status: string;
  count: number;
  percent: number;
  color: string;
}

export interface VMGroupStatus {
  groupName: string;
  count: number;
  online: number;
  offline: number;
  warning: number;
  critical: number;
}

export interface OSStatus {
  os: string;
  count: number;
  percent: number;
}

export interface VMStatusResponse {
  distribution: VMStatusDistribution[];
  byGroup: VMGroupStatus[];
  byOS: OSStatus[];
}

export interface DashboardAlert {
  id: string;
  vmId: string;
  vmName: string;
  vmIP: string;
  alertType: string;
  severity: string;
  message: string;
  value: string;
  threshold: string;
  occurredAt: string;
  status: string;
  acknowledged: boolean;
}

export interface AlertsResponse {
  alerts: DashboardAlert[];
  total: number;
  unreadCount: number;
}

export interface HealthTrendDataPoint {
  timestamp: string;
  score: number;
}

export interface HealthTrendResponse {
  period: string;
  currentScore: number;
  trend: string;
  dataPoints: HealthTrendDataPoint[];
}

export interface VMIssue {
  type: string;
  message: string;
  value: string;
}

export interface ProblemVM {
  vmId: string;
  vmName: string;
  vmIP: string;
  group: string;
  severity: string;
  issues: VMIssue[];
  firstDetected: string;
  duration: string;
}

export interface ProblemVMsResponse {
  total: number;
  vms: ProblemVM[];
}

const dashboardApi = {
  getOverview: async (): Promise<OverviewResponse> => {
    const response = await client.get<OverviewResponse>('/dashboard/overview');
    return response.data;
  },

  getVMStatus: async (): Promise<VMStatusResponse> => {
    const response = await client.get<VMStatusResponse>('/dashboard/vm-status');
    return response.data;
  },

  getAlerts: async (limit: number = 5): Promise<AlertsResponse> => {
    const response = await client.get<AlertsResponse>('/dashboard/alerts', {
      params: { limit },
    });
    return response.data;
  },

  getHealthTrend: async (period: string = '7d'): Promise<HealthTrendResponse> => {
    const response = await client.get<HealthTrendResponse>('/dashboard/health-trend', {
      params: { period },
    });
    return response.data;
  },

  getProblemVMs: async (severity?: string, limit: number = 20): Promise<ProblemVMsResponse> => {
    const response = await client.get<ProblemVMsResponse>('/dashboard/problem-vms', {
      params: { severity, limit },
    });
    return response.data;
  },
};

export default dashboardApi;
