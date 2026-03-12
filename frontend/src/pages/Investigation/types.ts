export interface MergedEvent {
  _time: string;
  _source_template: string; // 标记这条Data是哪个Rule查出来的 - ?: number; // 原始Rule - _rule_query?: string; // Rule的Query语句 - ?: string;
  action?: string;
  event_type?: string;
  severity?: string;
  raw_data?: string;
  [key: string]: any;
}

export interface InvestigationPageProps {
  tabData?: {
    incident_id?: number;
    case_id?: number;
    file_id?: number;
    params?: Record<string, string>;
  };
}