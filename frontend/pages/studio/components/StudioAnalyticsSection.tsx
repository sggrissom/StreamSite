import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import "./StudioAnalyticsSection-styles";

type StudioAnalyticsSectionProps = {
  studioId: number;
  studioName: string;
};

type State = {
  isLoading: boolean;
  error: string;
  analytics: server.GetStudioAnalyticsResponse | null;
  refreshInterval: number | null;
  cleanup: () => void;
};

const useState = vlens.declareHook((studioId: number): State => {
  const state: State = {
    isLoading: true,
    error: "",
    analytics: null,
    refreshInterval: null,
    cleanup: () => {
      if (state.refreshInterval !== null) {
        clearInterval(state.refreshInterval);
        state.refreshInterval = null;
      }
    },
  };

  // Initial fetch
  fetchAnalytics(state, studioId);

  // Set up auto-refresh every 30 seconds
  state.refreshInterval = window.setInterval(() => {
    fetchAnalytics(state, studioId);
  }, 30000);

  return state;
});

async function fetchAnalytics(state: State, studioId: number) {
  const [resp, err] = await server.GetStudioAnalytics({ studioId });

  if (err) {
    state.error = err || "Failed to load analytics";
    state.isLoading = false;
    vlens.scheduleRedraw();
    return;
  }

  state.analytics = resp;
  state.error = "";
  state.isLoading = false;
  vlens.scheduleRedraw();
}

export function StudioAnalyticsSection(props: StudioAnalyticsSectionProps) {
  const state = useState(props.studioId);

  return (
    <section className="studio-analytics-section">
      <div className="section-header">
        <h2>Studio Analytics</h2>
      </div>

      {state.isLoading && !state.analytics ? (
        <div className="analytics-loading">
          <p>Loading analytics...</p>
        </div>
      ) : state.error ? (
        <div className="analytics-error">
          <p>{state.error}</p>
        </div>
      ) : state.analytics && state.analytics.analytics ? (
        <div className="analytics-grid">
          {/* Current Viewers Card */}
          <div className="analytics-card">
            <div className="analytics-card-header">
              <div className="analytics-icon">👁️</div>
              <h3>Current Viewers</h3>
            </div>
            <div className="analytics-value analytics-value-primary">
              {state.analytics.analytics.currentViewers}
            </div>
            <div className="analytics-label">Watching now</div>
          </div>

          {/* Active Rooms Card */}
          <div className="analytics-card">
            <div className="analytics-card-header">
              <div className="analytics-icon">🔴</div>
              <h3>Active Streams</h3>
            </div>
            <div className="analytics-value">
              {state.analytics.analytics.activeRooms} /{" "}
              {state.analytics.analytics.totalRooms}
            </div>
            <div className="analytics-label">Rooms streaming</div>
          </div>

          {/* Total Views Card */}
          <div className="analytics-card">
            <div className="analytics-card-header">
              <div className="analytics-icon">📊</div>
              <h3>Total Views</h3>
            </div>
            <div className="analytics-value">
              {state.analytics.analytics.totalViewsAllTime.toLocaleString()}
            </div>
            <div className="analytics-sublabel">
              This month:{" "}
              {state.analytics.analytics.totalViewsThisMonth.toLocaleString()}
            </div>
          </div>

          {/* Total Stream Time Card */}
          <div className="analytics-card">
            <div className="analytics-card-header">
              <div className="analytics-icon">⏱️</div>
              <h3>Total Stream Time</h3>
            </div>
            <div className="analytics-value">
              {Math.floor(
                state.analytics.analytics.totalStreamMinutes / 60,
              ).toLocaleString()}{" "}
              <span className="analytics-unit">hours</span>
            </div>
            <div className="analytics-sublabel">
              {state.analytics.analytics.totalStreamMinutes % 60} minutes
            </div>
          </div>
        </div>
      ) : (
        <div className="analytics-empty">
          <p>No analytics data available yet.</p>
        </div>
      )}

      {state.analytics && (
        <div className="analytics-footer">
          <p className="analytics-refresh-note">
            Auto-refreshes every 30 seconds
          </p>
        </div>
      )}
    </section>
  );
}
