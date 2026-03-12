export interface MergedEvent {
  _time: string;
  _source_template: string; // 标记这条Data是哪个Rule查出来的
  [key: string]: any;
}

export interface InvestigationPageProps {
  tabData?: {
    incident_id?: number;
    params?: Record<string, string>;
  };
}