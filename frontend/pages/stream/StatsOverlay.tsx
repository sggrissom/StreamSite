import * as preact from "preact";
import "./stats-overlay-styles";

export type StreamStats = {
  currentResolution: string;
  currentBitrate: string;
  bandwidthEstimate: string;
  bufferHealth: string;
  isAutoQuality: boolean;
  droppedFrames: string;
  videoResolution: string;
  connectionTime: string;
  streamUptime: string;
};

type StatsOverlayProps = {
  stats: StreamStats | null;
  visible: boolean;
};

export function StatsOverlay(props: StatsOverlayProps) {
  if (!props.visible || !props.stats) return null;

  const { stats } = props;

  return (
    <div className="stats-overlay">
      <div className="stats-overlay-header">
        <span className="stats-overlay-title">Stream Stats</span>
        <span className="stats-overlay-hint">Press 'i' to hide</span>
      </div>
      <div className="stats-overlay-content">
        <div className="stats-row">
          <span className="stats-label">Resolution:</span>
          <span className="stats-value">
            {stats.currentResolution}
            {stats.isAutoQuality && <span className="stats-badge">Auto</span>}
          </span>
        </div>
        <div className="stats-row">
          <span className="stats-label">Bitrate:</span>
          <span className="stats-value">{stats.currentBitrate}</span>
        </div>
        <div className="stats-row">
          <span className="stats-label">Bandwidth:</span>
          <span className="stats-value">{stats.bandwidthEstimate}</span>
        </div>
        <div className="stats-row">
          <span className="stats-label">Buffer:</span>
          <span className="stats-value">{stats.bufferHealth}</span>
        </div>
        {stats.droppedFrames !== "N/A" && (
          <div className="stats-row">
            <span className="stats-label">Dropped Frames:</span>
            <span className="stats-value">{stats.droppedFrames}</span>
          </div>
        )}
        <div className="stats-row">
          <span className="stats-label">Video Size:</span>
          <span className="stats-value">{stats.videoResolution}</span>
        </div>
        <div className="stats-row">
          <span className="stats-label">Watching:</span>
          <span className="stats-value">{stats.connectionTime}</span>
        </div>
        <div className="stats-row">
          <span className="stats-label">Stream Uptime:</span>
          <span className="stats-value">{stats.streamUptime}</span>
        </div>
      </div>
    </div>
  );
}
