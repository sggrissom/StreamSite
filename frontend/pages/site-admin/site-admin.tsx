import * as preact from "preact";
import * as vlens from "vlens";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import { Modal } from "../../components/Modal";
import "../../styles/global";
import "./site-admin-styles";

type Data = server.ListAllStudiosResponse;

export async function fetch(route: string, prefix: string) {
  return server.ListAllStudios({});
}

// ===== Delete Studio Modal =====
type DeleteStudioModal = {
  isOpen: boolean;
  isDeleting: boolean;
  error: string;
  studioId: number;
  studioName: string;
};

const useDeleteStudioModal = vlens.declareHook(
  (): DeleteStudioModal => ({
    isOpen: false,
    isDeleting: false,
    error: "",
    studioId: 0,
    studioName: "",
  }),
);

function openDeleteStudioModal(
  modal: DeleteStudioModal,
  studioId: number,
  studioName: string,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.studioId = studioId;
  modal.studioName = studioName;
  vlens.scheduleRedraw();
}

function closeDeleteStudioModal(modal: DeleteStudioModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function confirmDeleteStudio(modal: DeleteStudioModal) {
  modal.isDeleting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.DeleteStudio({
    studioId: modal.studioId,
  });

  modal.isDeleting = false;

  if (err) {
    modal.error = err || "Failed to delete studio";
    vlens.scheduleRedraw();
    return;
  }

  closeDeleteStudioModal(modal);
  window.location.reload();
}

// ===== Page Navigation State =====
type PageState = {
  activeSection: "studios" | "users" | "logs" | "performance";
};

const usePageState = vlens.declareHook(
  (): PageState => ({
    activeSection: "studios",
  }),
);

function setActiveSection(
  state: PageState,
  section: "studios" | "users" | "logs" | "performance",
) {
  state.activeSection = section;
  vlens.scheduleRedraw();
}

// ===== Users Management =====
type UsersState = {
  users: server.UserWithStats[];
  loading: boolean;
  error: string;
  changingRoleFor: number; // userId currently being updated
};

const useUsersState = vlens.declareHook(
  (): UsersState => ({
    users: [],
    loading: false,
    error: "",
    changingRoleFor: 0,
  }),
);

async function loadUsers(state: UsersState) {
  state.loading = true;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.ListAllUsers({});

  state.loading = false;

  if (err || !resp) {
    state.error = err || "Failed to load users";
    vlens.scheduleRedraw();
    return;
  }

  state.users = resp.users || [];
  vlens.scheduleRedraw();
}

async function changeUserRole(
  state: UsersState,
  userId: number,
  newRole: number,
  myUserId: number,
) {
  if (userId === myUserId) {
    return; // Cannot change own role
  }

  state.changingRoleFor = userId;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.UpdateUserRole({
    userId: userId,
    newRole: newRole,
  });

  state.changingRoleFor = 0;

  if (err) {
    state.error = err || "Failed to update user role";
    vlens.scheduleRedraw();
    return;
  }

  // Reload users to get fresh data
  await loadUsers(state);
}

function getRoleName(role: number): string {
  switch (role) {
    case 0:
      return "User";
    case 1:
      return "Stream Admin";
    case 2:
      return "Site Admin";
    default:
      return "Unknown";
  }
}

// ===== Logs Management =====
type LogsState = {
  logs: server.LogEntry[];
  loading: boolean;
  error: string;
  totalCount: number;
  filterLevel: string;
  filterCategory: string;
  filterSearch: string;
  expandedLogIndex: number; // -1 means none expanded
};

const useLogsState = vlens.declareHook(
  (): LogsState => ({
    logs: [],
    loading: false,
    error: "",
    totalCount: 0,
    filterLevel: "",
    filterCategory: "",
    filterSearch: "",
    expandedLogIndex: -1,
  }),
);

async function loadLogs(state: LogsState) {
  state.loading = true;
  state.error = "";
  vlens.scheduleRedraw();

  const req: server.GetSystemLogsRequest = {
    level: state.filterLevel || null,
    category: state.filterCategory || null,
    search: state.filterSearch || null,
  };

  const [resp, err] = await server.GetSystemLogs(req);

  state.loading = false;

  if (err || !resp) {
    state.error = err || "Failed to load logs";
    vlens.scheduleRedraw();
    return;
  }

  state.logs = resp.logs || [];
  state.totalCount = resp.totalCount;
  vlens.scheduleRedraw();
}

function toggleLogExpansion(state: LogsState, index: number) {
  state.expandedLogIndex = state.expandedLogIndex === index ? -1 : index;
  vlens.scheduleRedraw();
}

function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    return date.toLocaleString();
  } catch {
    return timestamp;
  }
}

function getLevelClass(level: string): string {
  switch (level) {
    case "INFO":
      return "log-level-info";
    case "WARN":
      return "log-level-warn";
    case "ERROR":
      return "log-level-error";
    case "DEBUG":
      return "log-level-debug";
    default:
      return "log-level-unknown";
  }
}

// ===== Performance Metrics =====
type PerformanceState = {
  siteWide: server.SitePerformanceMetrics | null;
  perStudio: server.StudioPerformanceMetrics[];
  loading: boolean;
  error: string;
  sortColumn: "studioName" | "ttff" | "rebuffer" | "bitrate" | "errorRate";
  sortDirection: "asc" | "desc";
};

const usePerformanceState = vlens.declareHook(
  (): PerformanceState => ({
    siteWide: null,
    perStudio: [],
    loading: false,
    error: "",
    sortColumn: "studioName",
    sortDirection: "asc",
  }),
);

async function loadPerformanceMetrics(state: PerformanceState) {
  state.loading = true;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.GetSitePerformanceMetrics({});

  state.loading = false;

  if (err || !resp) {
    state.error = err || "Failed to load performance metrics";
    vlens.scheduleRedraw();
    return;
  }

  state.siteWide = resp.siteWide;
  state.perStudio = resp.perStudio || [];
  vlens.scheduleRedraw();
}

function setSortColumn(
  state: PerformanceState,
  column: "studioName" | "ttff" | "rebuffer" | "bitrate" | "errorRate",
) {
  if (state.sortColumn === column) {
    state.sortDirection = state.sortDirection === "asc" ? "desc" : "asc";
  } else {
    state.sortColumn = column;
    state.sortDirection = "asc";
  }
  vlens.scheduleRedraw();
}

function getSortedStudios(
  state: PerformanceState,
): server.StudioPerformanceMetrics[] {
  const studios = [...state.perStudio];
  studios.sort((a, b) => {
    let aVal: number | string;
    let bVal: number | string;

    switch (state.sortColumn) {
      case "studioName":
        aVal = a.studioName.toLowerCase();
        bVal = b.studioName.toLowerCase();
        break;
      case "ttff":
        aVal = a.avgTimeToFirstFrame;
        bVal = b.avgTimeToFirstFrame;
        break;
      case "rebuffer":
        aVal = a.avgRebufferRatio;
        bVal = b.avgRebufferRatio;
        break;
      case "bitrate":
        aVal = a.avgBitrateMbps;
        bVal = b.avgBitrateMbps;
        break;
      case "errorRate":
        aVal = a.errorRate;
        bVal = b.errorRate;
        break;
      default:
        return 0;
    }

    if (aVal < bVal) return state.sortDirection === "asc" ? -1 : 1;
    if (aVal > bVal) return state.sortDirection === "asc" ? 1 : -1;
    return 0;
  });

  return studios;
}

function formatPercentage(value: number): string {
  return value.toFixed(2) + "%";
}

function formatMilliseconds(value: number): string {
  return value.toFixed(0) + "ms";
}

function formatMbps(value: number): string {
  return value.toFixed(2) + " Mbps";
}

function getPerformanceClass(
  metric: "ttff" | "rebuffer" | "errorRate",
  value: number,
): string {
  switch (metric) {
    case "ttff":
      if (value < 1000) return "perf-good";
      if (value < 2000) return "perf-warning";
      return "perf-bad";
    case "rebuffer":
      if (value < 1) return "perf-good";
      if (value < 3) return "perf-warning";
      return "perf-bad";
    case "errorRate":
      if (value < 1) return "perf-good";
      if (value < 5) return "perf-warning";
      return "perf-bad";
    default:
      return "";
  }
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const deleteStudioModal = useDeleteStudioModal();
  const usersState = useUsersState();
  const performanceState = usePerformanceState();
  const logsState = useLogsState();
  const pageState = usePageState();

  // Check permissions
  if (!data || !data.studios) {
    return (
      <div>
        <Header />
        <main className="site-admin-container">
          <div className="site-admin-content">
            <div className="error-state">
              <div className="error-icon">⚠️</div>
              <h2>Access Denied</h2>
              <p>
                You don't have permission to access this page. Only site
                administrators can view this area.
              </p>
              <a href="/dashboard" className="btn btn-primary">
                Back to Dashboard
              </a>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const studios = data.studios || [];

  // Load users on first render
  if (usersState.users.length === 0 && !usersState.loading) {
    loadUsers(usersState);
  }

  return (
    <div>
      <Header />
      <main className="site-admin-container">
        <div className="site-admin-content">
          <div className="site-admin-header">
            <h1 className="site-admin-title">Site Administration</h1>
            <p className="site-admin-description">
              Manage all studios, users, and system-wide settings
            </p>
          </div>

          {/* Navigation Tabs */}
          <div className="section-nav">
            <button
              className={`nav-tab ${pageState.activeSection === "studios" ? "active" : ""}`}
              onClick={() => setActiveSection(pageState, "studios")}
            >
              Studios
            </button>
            <button
              className={`nav-tab ${pageState.activeSection === "users" ? "active" : ""}`}
              onClick={() => setActiveSection(pageState, "users")}
            >
              Users
            </button>
            <button
              className={`nav-tab ${pageState.activeSection === "logs" ? "active" : ""}`}
              onClick={() => setActiveSection(pageState, "logs")}
            >
              System Logs
            </button>
            <button
              className={`nav-tab ${pageState.activeSection === "performance" ? "active" : ""}`}
              onClick={() => setActiveSection(pageState, "performance")}
            >
              Performance
            </button>
          </div>

          {/* Studios Management Section */}
          {pageState.activeSection === "studios" && (
            <div className="admin-section">
              <div className="section-header">
                <h2 className="section-title">Studios Management</h2>
                <a href="/studios" className="btn btn-primary btn-sm">
                  Create Studio
                </a>
              </div>

              {studios.length === 0 ? (
                <div className="empty-state">
                  <p>No studios have been created yet.</p>
                </div>
              ) : (
                <div className="studios-table-container">
                  <table className="admin-table">
                    <thead>
                      <tr>
                        <th>Studio Name</th>
                        <th>Owner</th>
                        <th>Created</th>
                        <th>Rooms</th>
                        <th>Members</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {studios.map((studio) => (
                        <tr key={studio.id}>
                          <td>
                            <div className="studio-name-cell">
                              <strong>{studio.name}</strong>
                              {studio.description && (
                                <div className="studio-description-preview">
                                  {studio.description}
                                </div>
                              )}
                            </div>
                          </td>
                          <td>
                            <div className="owner-cell">
                              <div>{studio.ownerName}</div>
                              <div className="email-text">
                                {studio.ownerEmail}
                              </div>
                            </div>
                          </td>
                          <td>
                            {new Date(studio.creation).toLocaleDateString()}
                          </td>
                          <td>{studio.roomCount}</td>
                          <td>{studio.memberCount}</td>
                          <td className="actions-cell">
                            <a
                              href={`/studio/${studio.id}`}
                              className="btn btn-secondary btn-sm"
                            >
                              View Studio
                            </a>
                            <button
                              className="btn btn-danger btn-sm"
                              onClick={() =>
                                openDeleteStudioModal(
                                  deleteStudioModal,
                                  studio.id,
                                  studio.name,
                                )
                              }
                            >
                              Delete
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* Users Management Section */}
          {pageState.activeSection === "users" && (
            <div className="admin-section">
              <div className="section-header">
                <h2 className="section-title">Users Management</h2>
              </div>

              {usersState.error && (
                <div className="error-message">{usersState.error}</div>
              )}

              {usersState.loading ? (
                <div className="loading-state">Loading users...</div>
              ) : usersState.users.length === 0 ? (
                <div className="empty-state">
                  <p>No users found.</p>
                </div>
              ) : (
                <div className="users-table-container">
                  <table className="admin-table">
                    <thead>
                      <tr>
                        <th>Name</th>
                        <th>Email</th>
                        <th>Role</th>
                        <th>Studios</th>
                        <th>Created</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {usersState.users.map((user) => (
                        <tr key={user.id}>
                          <td>{user.name}</td>
                          <td>{user.email}</td>
                          <td>
                            <span className={`role-badge role-${user.role}`}>
                              {getRoleName(user.role)}
                            </span>
                          </td>
                          <td>{user.studioCount}</td>
                          <td>
                            {new Date(user.creation).toLocaleDateString()}
                          </td>
                          <td className="actions-cell">
                            {/* Get current user ID from first user (yourself) - you're always the site admin loading this */}
                            {user.id === 1 ? (
                              <span className="text-muted">(You)</span>
                            ) : (
                              <select
                                className="role-select"
                                disabled={
                                  usersState.changingRoleFor === user.id
                                }
                                onChange={(e) => {
                                  const target = e.target as HTMLSelectElement;
                                  changeUserRole(
                                    usersState,
                                    user.id,
                                    parseInt(target.value),
                                    1, // Site admin is always user ID 1
                                  );
                                }}
                                value={user.role}
                              >
                                <option value={0}>User</option>
                                <option value={1}>Stream Admin</option>
                                <option value={2}>Site Admin</option>
                              </select>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* Logs Management Section */}
          {pageState.activeSection === "logs" && (
            <div className="admin-section">
              <div className="section-header">
                <h2 className="section-title">System Logs</h2>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => loadLogs(logsState)}
                  disabled={logsState.loading}
                >
                  {logsState.loading ? "Refreshing..." : "Refresh Logs"}
                </button>
              </div>

              {/* Filter Controls */}
              <div className="log-filters">
                <div className="filter-row">
                  <div className="filter-group">
                    <label htmlFor="level-filter">Level:</label>
                    <select
                      id="level-filter"
                      className="filter-select"
                      value={logsState.filterLevel}
                      onChange={(e) => {
                        const target = e.target as HTMLSelectElement;
                        logsState.filterLevel = target.value;
                        vlens.scheduleRedraw();
                      }}
                    >
                      <option value="">All</option>
                      <option value="INFO">INFO</option>
                      <option value="WARN">WARN</option>
                      <option value="ERROR">ERROR</option>
                      <option value="DEBUG">DEBUG</option>
                    </select>
                  </div>

                  <div className="filter-group">
                    <label htmlFor="category-filter">Category:</label>
                    <select
                      id="category-filter"
                      className="filter-select"
                      value={logsState.filterCategory}
                      onChange={(e) => {
                        const target = e.target as HTMLSelectElement;
                        logsState.filterCategory = target.value;
                        vlens.scheduleRedraw();
                      }}
                    >
                      <option value="">All</option>
                      <option value="AUTH">AUTH</option>
                      <option value="STREAM">STREAM</option>
                      <option value="API">API</option>
                      <option value="SYSTEM">SYSTEM</option>
                    </select>
                  </div>

                  <div className="filter-group filter-search-group">
                    <label htmlFor="search-filter">Search:</label>
                    <input
                      id="search-filter"
                      type="text"
                      className="filter-input"
                      placeholder="Search in messages..."
                      value={logsState.filterSearch}
                      onInput={(e) => {
                        const target = e.target as HTMLInputElement;
                        logsState.filterSearch = target.value;
                        vlens.scheduleRedraw();
                      }}
                    />
                  </div>

                  <button
                    className="btn btn-secondary btn-sm"
                    onClick={() => loadLogs(logsState)}
                    disabled={logsState.loading}
                  >
                    Apply Filters
                  </button>
                </div>

                {logsState.totalCount > 0 && (
                  <div className="log-count">
                    Showing {logsState.logs.length} of {logsState.totalCount}{" "}
                    logs
                  </div>
                )}
              </div>

              {logsState.error && (
                <div className="error-message">{logsState.error}</div>
              )}

              {logsState.loading ? (
                <div className="loading-state">Loading logs...</div>
              ) : logsState.logs.length === 0 ? (
                <div className="empty-state">
                  <p>No logs found. Try adjusting your filters.</p>
                </div>
              ) : (
                <div className="logs-table-container">
                  <table className="admin-table logs-table">
                    <thead>
                      <tr>
                        <th>Timestamp</th>
                        <th>Level</th>
                        <th>Category</th>
                        <th>Message</th>
                        <th>Details</th>
                      </tr>
                    </thead>
                    <tbody>
                      {logsState.logs.map((log, index) => (
                        <>
                          <tr key={index} className="log-row">
                            <td className="log-timestamp">
                              {formatTimestamp(log.timestamp)}
                            </td>
                            <td>
                              <span
                                className={`log-level-badge ${getLevelClass(log.level)}`}
                              >
                                {log.level}
                              </span>
                            </td>
                            <td>
                              <span className="log-category">
                                {log.category}
                              </span>
                            </td>
                            <td className="log-message">{log.message}</td>
                            <td className="actions-cell">
                              {(log.data ||
                                log.userId ||
                                log.ip ||
                                log.userAgent) && (
                                <button
                                  className="btn btn-secondary btn-sm"
                                  onClick={() =>
                                    toggleLogExpansion(logsState, index)
                                  }
                                >
                                  {logsState.expandedLogIndex === index
                                    ? "Hide"
                                    : "Details"}
                                </button>
                              )}
                            </td>
                          </tr>
                          {logsState.expandedLogIndex === index && (
                            <tr
                              key={`${index}-details`}
                              className="log-details-row"
                            >
                              <td colSpan={5}>
                                <div className="log-details">
                                  {log.userId && (
                                    <div className="log-detail-item">
                                      <strong>User ID:</strong> {log.userId}
                                    </div>
                                  )}
                                  {log.ip && (
                                    <div className="log-detail-item">
                                      <strong>IP:</strong> {log.ip}
                                    </div>
                                  )}
                                  {log.userAgent && (
                                    <div className="log-detail-item">
                                      <strong>User Agent:</strong>{" "}
                                      {log.userAgent}
                                    </div>
                                  )}
                                  {log.data &&
                                    Object.keys(log.data).length > 0 && (
                                      <div className="log-detail-item">
                                        <strong>Data:</strong>
                                        <pre className="log-data-json">
                                          {JSON.stringify(log.data, null, 2)}
                                        </pre>
                                      </div>
                                    )}
                                </div>
                              </td>
                            </tr>
                          )}
                        </>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* Performance Metrics Section */}
          {pageState.activeSection === "performance" && (
            <div className="admin-section">
              <div className="section-header">
                <h2 className="section-title">Site Performance Metrics</h2>
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => loadPerformanceMetrics(performanceState)}
                  disabled={performanceState.loading}
                >
                  {performanceState.loading ? "Loading..." : "Refresh Metrics"}
                </button>
              </div>

              {performanceState.error && (
                <div className="error-state">{performanceState.error}</div>
              )}

              {!performanceState.loading && performanceState.siteWide && (
                <div>
                  {/* Site-Wide Metrics Grid */}
                  <div className="perf-metrics-grid">
                    <div className="perf-metric-card">
                      <div className="metric-label">
                        Avg Time to First Frame
                      </div>
                      <div
                        className={`metric-value ${getPerformanceClass("ttff", performanceState.siteWide.avgTimeToFirstFrame)}`}
                      >
                        {formatMilliseconds(
                          performanceState.siteWide.avgTimeToFirstFrame,
                        )}
                      </div>
                      <div className="metric-context">
                        {performanceState.siteWide.totalStartupAttempts}{" "}
                        attempts
                      </div>
                    </div>

                    <div className="perf-metric-card">
                      <div className="metric-label">Startup Success Rate</div>
                      <div className="metric-value">
                        {formatPercentage(
                          performanceState.siteWide.startupSuccessRate,
                        )}
                      </div>
                      <div className="metric-context">
                        {performanceState.siteWide.totalStartupFailures}{" "}
                        failures
                      </div>
                    </div>

                    <div className="perf-metric-card">
                      <div className="metric-label">Avg Rebuffer Ratio</div>
                      <div
                        className={`metric-value ${getPerformanceClass("rebuffer", performanceState.siteWide.avgRebufferRatio)}`}
                      >
                        {formatPercentage(
                          performanceState.siteWide.avgRebufferRatio,
                        )}
                      </div>
                      <div className="metric-context">
                        {performanceState.siteWide.totalRebufferEvents} events
                      </div>
                    </div>

                    <div className="perf-metric-card">
                      <div className="metric-label">Avg Bitrate</div>
                      <div className="metric-value">
                        {formatMbps(performanceState.siteWide.avgBitrateMbps)}
                      </div>
                      <div className="metric-context">across all streams</div>
                    </div>

                    <div className="perf-metric-card">
                      <div className="metric-label">Error Rate</div>
                      <div
                        className={`metric-value ${getPerformanceClass("errorRate", performanceState.siteWide.errorRate)}`}
                      >
                        {formatPercentage(performanceState.siteWide.errorRate)}
                      </div>
                      <div className="metric-context">
                        {performanceState.siteWide.totalErrors} errors total
                      </div>
                    </div>

                    <div className="perf-metric-card">
                      <div className="metric-label">Quality Distribution</div>
                      <div className="quality-bar">
                        {performanceState.siteWide.quality480pPercent > 0 && (
                          <div
                            className="quality-segment q-480p"
                            style={`width: ${performanceState.siteWide.quality480pPercent}%`}
                            title={`480p: ${formatPercentage(performanceState.siteWide.quality480pPercent)}`}
                          />
                        )}
                        {performanceState.siteWide.quality720pPercent > 0 && (
                          <div
                            className="quality-segment q-720p"
                            style={`width: ${performanceState.siteWide.quality720pPercent}%`}
                            title={`720p: ${formatPercentage(performanceState.siteWide.quality720pPercent)}`}
                          />
                        )}
                        {performanceState.siteWide.quality1080pPercent > 0 && (
                          <div
                            className="quality-segment q-1080p"
                            style={`width: ${performanceState.siteWide.quality1080pPercent}%`}
                            title={`1080p: ${formatPercentage(performanceState.siteWide.quality1080pPercent)}`}
                          />
                        )}
                      </div>
                      <div className="metric-context quality-legend">
                        <span className="legend-item">
                          <span className="legend-color q-480p"></span> 480p:{" "}
                          {formatPercentage(
                            performanceState.siteWide.quality480pPercent,
                          )}
                        </span>
                        <span className="legend-item">
                          <span className="legend-color q-720p"></span> 720p:{" "}
                          {formatPercentage(
                            performanceState.siteWide.quality720pPercent,
                          )}
                        </span>
                        <span className="legend-item">
                          <span className="legend-color q-1080p"></span> 1080p:{" "}
                          {formatPercentage(
                            performanceState.siteWide.quality1080pPercent,
                          )}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Per-Studio Performance Table */}
                  {performanceState.perStudio.length > 0 && (
                    <div className="perf-table-section">
                      <h3 className="subsection-title">
                        Performance by Studio
                      </h3>
                      <div className="table-responsive">
                        <table className="admin-table">
                          <thead>
                            <tr>
                              <th
                                className="sortable"
                                onClick={() =>
                                  setSortColumn(performanceState, "studioName")
                                }
                              >
                                Studio{" "}
                                {performanceState.sortColumn === "studioName" &&
                                  (performanceState.sortDirection === "asc"
                                    ? "↑"
                                    : "↓")}
                              </th>
                              <th>Rooms</th>
                              <th
                                className="sortable"
                                onClick={() =>
                                  setSortColumn(performanceState, "ttff")
                                }
                              >
                                Avg TTFF{" "}
                                {performanceState.sortColumn === "ttff" &&
                                  (performanceState.sortDirection === "asc"
                                    ? "↑"
                                    : "↓")}
                              </th>
                              <th>Success Rate</th>
                              <th
                                className="sortable"
                                onClick={() =>
                                  setSortColumn(performanceState, "rebuffer")
                                }
                              >
                                Rebuffer Ratio{" "}
                                {performanceState.sortColumn === "rebuffer" &&
                                  (performanceState.sortDirection === "asc"
                                    ? "↑"
                                    : "↓")}
                              </th>
                              <th
                                className="sortable"
                                onClick={() =>
                                  setSortColumn(performanceState, "bitrate")
                                }
                              >
                                Avg Bitrate{" "}
                                {performanceState.sortColumn === "bitrate" &&
                                  (performanceState.sortDirection === "asc"
                                    ? "↑"
                                    : "↓")}
                              </th>
                              <th
                                className="sortable"
                                onClick={() =>
                                  setSortColumn(performanceState, "errorRate")
                                }
                              >
                                Error Rate{" "}
                                {performanceState.sortColumn === "errorRate" &&
                                  (performanceState.sortDirection === "asc"
                                    ? "↑"
                                    : "↓")}
                              </th>
                              <th>Total Errors</th>
                            </tr>
                          </thead>
                          <tbody>
                            {getSortedStudios(performanceState).map(
                              (studio) => (
                                <tr key={studio.studioId}>
                                  <td>
                                    <a
                                      href={`/studio/${studio.studioId}`}
                                      className="studio-link"
                                    >
                                      {studio.studioName}
                                    </a>
                                  </td>
                                  <td>{studio.totalRooms}</td>
                                  <td
                                    className={getPerformanceClass(
                                      "ttff",
                                      studio.avgTimeToFirstFrame,
                                    )}
                                  >
                                    {formatMilliseconds(
                                      studio.avgTimeToFirstFrame,
                                    )}
                                  </td>
                                  <td>
                                    {formatPercentage(
                                      studio.startupSuccessRate,
                                    )}
                                  </td>
                                  <td
                                    className={getPerformanceClass(
                                      "rebuffer",
                                      studio.avgRebufferRatio,
                                    )}
                                  >
                                    {formatPercentage(studio.avgRebufferRatio)}
                                  </td>
                                  <td>{formatMbps(studio.avgBitrateMbps)}</td>
                                  <td
                                    className={getPerformanceClass(
                                      "errorRate",
                                      studio.errorRate,
                                    )}
                                  >
                                    {formatPercentage(studio.errorRate)}
                                  </td>
                                  <td>{studio.totalErrors}</td>
                                </tr>
                              ),
                            )}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {performanceState.perStudio.length === 0 &&
                    !performanceState.loading && (
                      <div className="empty-state">
                        <p>No performance data available yet.</p>
                        <p className="empty-state-hint">
                          Performance metrics will appear once viewers start
                          watching streams.
                        </p>
                      </div>
                    )}
                </div>
              )}

              {performanceState.loading && (
                <div className="loading-state">
                  Loading performance metrics...
                </div>
              )}
            </div>
          )}
        </div>
      </main>

      {/* Delete Studio Modal */}
      <Modal
        isOpen={deleteStudioModal.isOpen}
        title="Delete Studio"
        onClose={() => closeDeleteStudioModal(deleteStudioModal)}
        error={deleteStudioModal.error}
        footer={
          <>
            <button
              className="btn btn-secondary"
              onClick={() => closeDeleteStudioModal(deleteStudioModal)}
              disabled={deleteStudioModal.isDeleting}
            >
              Cancel
            </button>
            <button
              className="btn btn-danger"
              onClick={() => confirmDeleteStudio(deleteStudioModal)}
              disabled={deleteStudioModal.isDeleting}
            >
              {deleteStudioModal.isDeleting ? "Deleting..." : "Delete Studio"}
            </button>
          </>
        }
      >
        <div className="delete-confirmation">
          <p className="confirmation-text">
            ⚠️ Are you sure you want to delete this studio? This will also
            delete all rooms, streams, and memberships. This action cannot be
            undone.
          </p>
          <div className="studio-info">
            <strong>Studio:</strong> {deleteStudioModal.studioName}
          </div>
        </div>
      </Modal>

      <Footer />
    </div>
  );
}
