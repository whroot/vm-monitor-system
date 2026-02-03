
export type Language = 'en' | 'zh' | 'jp';

export enum VMStatus {
  ONLINE = 'online',
  OFFLINE = 'offline',
  WARNING = 'warning',
  CRITICAL = 'critical'
}

export interface VMData {
  id: string;
  name: string;
  ip: string;
  status: VMStatus;
  cpu: number;
  memory: number;
  disk: number;
  network: number;
  os: string;
  uptime: string;
}

export interface Alert {
  id: string;
  time: string;
  vmName: string;
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  message: string;
  status: 'pending' | 'resolved' | 'ignored';
}

export type DashboardMode = 'normal' | 'fault';

export interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  status: 'active' | 'pending' | 'disabled' | 'expired';
  lastLogin: string;
}

export interface Permission {
  module: string;
  level: 'none' | 'read' | 'edit' | 'admin';
  source: 'direct' | 'inherited';
}
