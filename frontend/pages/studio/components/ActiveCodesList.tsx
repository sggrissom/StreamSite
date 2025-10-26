import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";

type Room = {
  id: number;
  name: string;
};

type ActiveCodesListProps = {
  studioId: number;
  studioName: string;
  rooms: Room[];
};

// ===== Component State =====
type CodeWithScope = server.AccessCodeListItem & {
  scopeName: string; // Room name or "Studio-wide"
};

type ActiveCodesListState = {
  codes: CodeWithScope[];
  isLoading: boolean;
  error: string;
  filterShowAll: boolean;
  refreshTimerId: number | null;
  hasInitiallyLoaded: boolean;
  revokeModal: RevokeModalState;
};

type RevokeModalState = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  code: string;
  label: string;
};

const useActiveCodesListState = vlens.declareHook(
  (): ActiveCodesListState => ({
    codes: [],
    isLoading: false,
    error: "",
    filterShowAll: false,
    refreshTimerId: null,
    hasInitiallyLoaded: false,
    revokeModal: {
      isOpen: false,
      isSubmitting: false,
      error: "",
      code: "",
      label: "",
    },
  }),
);

// ===== Helper Functions =====
async function loadCodes(
  state: ActiveCodesListState,
  studioId: number,
  studioName: string,
  rooms: Room[],
) {
  state.isLoading = true;
  state.error = "";
  vlens.scheduleRedraw();

  // Fetch codes from all rooms in parallel
  const roomCodePromises = rooms.map(async (room) => {
    const [resp, err] = await server.ListAccessCodes({
      type: 0, // Room codes
      targetId: room.id,
    });
    if (err || !resp) {
      return { codes: [], roomName: room.name };
    }
    return { codes: resp.codes || [], roomName: room.name };
  });

  // Fetch studio-wide codes
  const studioCodesPromise = server.ListAccessCodes({
    type: 1, // Studio codes
    targetId: studioId,
  });

  // Wait for all requests
  const roomResults = await Promise.all(roomCodePromises);
  const [studioResp, studioErr] = await studioCodesPromise;

  state.isLoading = false;

  // Merge all codes with scope information
  const allCodes: CodeWithScope[] = [];

  // Add room codes
  for (const result of roomResults) {
    for (const code of result.codes) {
      allCodes.push({
        ...code,
        scopeName: result.roomName,
      });
    }
  }

  // Add studio codes
  if (studioResp && studioResp.codes) {
    for (const code of studioResp.codes || []) {
      allCodes.push({
        ...code,
        scopeName: "Studio-wide",
      });
    }
  }

  // Sort by creation time (newest first)
  allCodes.sort((a, b) => {
    const dateA = new Date(a.createdAt).getTime();
    const dateB = new Date(b.createdAt).getTime();
    return dateB - dateA;
  });

  state.codes = allCodes;
  vlens.scheduleRedraw();
}

function openRevokeModal(
  state: ActiveCodesListState,
  code: string,
  label: string,
) {
  state.revokeModal.isOpen = true;
  state.revokeModal.code = code;
  state.revokeModal.label = label;
  state.revokeModal.error = "";
  vlens.scheduleRedraw();
}

function closeRevokeModal(state: ActiveCodesListState) {
  state.revokeModal.isOpen = false;
  state.revokeModal.error = "";
  vlens.scheduleRedraw();
}

async function submitRevoke(
  state: ActiveCodesListState,
  studioId: number,
  studioName: string,
  rooms: Room[],
) {
  state.revokeModal.isSubmitting = true;
  state.revokeModal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.RevokeAccessCode({
    code: state.revokeModal.code,
  });

  state.revokeModal.isSubmitting = false;

  if (err || !resp) {
    state.revokeModal.error = err || "Failed to revoke code";
    vlens.scheduleRedraw();
    return;
  }

  // Success - close modal and reload codes
  closeRevokeModal(state);
  loadCodes(state, studioId, studioName, rooms);
}

function toggleFilter(state: ActiveCodesListState) {
  state.filterShowAll = !state.filterShowAll;
  vlens.scheduleRedraw();
}

function formatTimeRemaining(expiresAt: string): string {
  const now = new Date().getTime();
  const expires = new Date(expiresAt).getTime();
  const diff = expires - now;

  if (diff <= 0) return "Expired";

  // Check if expiration is more than 50 years away (never expires)
  const fiftyYearsInMs = 50 * 365.25 * 24 * 60 * 60 * 1000;
  if (diff > fiftyYearsInMs) return "Never expires";

  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

function getStatusBadge(
  code: server.AccessCodeListItem,
): preact.ComponentChild {
  if (code.isRevoked) {
    return <span className="status-badge status-revoked">Revoked</span>;
  }
  if (code.isExpired) {
    return <span className="status-badge status-expired">Expired</span>;
  }
  return <span className="status-badge status-active">Active</span>;
}

function getStatusClass(code: server.AccessCodeListItem): string {
  if (code.isRevoked) return "status-revoked";
  if (code.isExpired) return "status-expired";
  return "status-active";
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text);
}

// ===== Main Component =====
export function ActiveCodesList(props: ActiveCodesListProps) {
  const state = useActiveCodesListState();

  // Initial load - only trigger once
  if (!state.hasInitiallyLoaded && !state.isLoading) {
    state.hasInitiallyLoaded = true;
    loadCodes(state, props.studioId, props.studioName, props.rooms);
  }

  // Setup auto-refresh timer (30 seconds)
  if (state.refreshTimerId === null) {
    const timerId = window.setInterval(() => {
      loadCodes(state, props.studioId, props.studioName, props.rooms);
    }, 30000);
    state.refreshTimerId = timerId;
  }

  // Cleanup timer on unmount would require a different pattern in vlens
  // For now, the timer will refresh as long as the component is mounted

  // Filter codes based on filter setting
  const displayCodes = state.filterShowAll
    ? state.codes
    : state.codes.filter((c) => !c.isExpired && !c.isRevoked);

  return (
    <div className="access-codes-section">
      <div className="section-header">
        <h2>Access Codes</h2>
        <div className="section-actions">
          <button
            className="btn btn-secondary btn-sm"
            onClick={() => toggleFilter(state)}
          >
            {state.filterShowAll ? "Show Active Only" : "Show All"}
          </button>
          <button
            className="btn btn-secondary btn-sm"
            onClick={() =>
              loadCodes(state, props.studioId, props.studioName, props.rooms)
            }
            disabled={state.isLoading}
          >
            {state.isLoading ? "Refreshing..." : "Refresh"}
          </button>
        </div>
      </div>

      {state.error && (
        <div className="error-message">
          <span>⚠️ {state.error}</span>
        </div>
      )}

      {state.isLoading && state.codes.length === 0 && (
        <div className="loading-state">Loading codes...</div>
      )}

      {!state.isLoading && state.codes.length === 0 && !state.error && (
        <div className="empty-state">
          <p>
            No access codes yet. Generate one from the room actions menu above.
          </p>
        </div>
      )}

      {displayCodes.length === 0 && state.codes.length > 0 && (
        <div className="empty-state">
          <p>No active codes. Click "Show All" to see expired/revoked codes.</p>
        </div>
      )}

      {displayCodes.length > 0 && (
        <div className="codes-table-container">
          <table className="codes-table">
            <thead>
              <tr>
                <th>Code</th>
                <th>Label</th>
                <th>Scope</th>
                <th>Status</th>
                <th>Expires</th>
                <th>Viewers</th>
                <th>Total</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {displayCodes.map((code) => (
                <tr key={code.code} className={getStatusClass(code)}>
                  <td className="code-cell">
                    <span className="code-display">{code.code}</span>
                    <button
                      className="btn-icon"
                      onClick={() => copyToClipboard(code.code)}
                      title="Copy code"
                    >
                      📋
                    </button>
                  </td>
                  <td className="label-cell">
                    {code.label || <em>No label</em>}
                  </td>
                  <td className="scope-cell">
                    <span
                      className={
                        code.scopeName === "Studio-wide"
                          ? "scope-studio"
                          : "scope-room"
                      }
                    >
                      {code.scopeName}
                    </span>
                  </td>
                  <td className="status-cell">{getStatusBadge(code)}</td>
                  <td className="expires-cell">
                    {code.isExpired ? (
                      <span className="expired-text">Expired</span>
                    ) : (
                      <span className="countdown">
                        {formatTimeRemaining(code.expiresAt)}
                      </span>
                    )}
                  </td>
                  <td className="viewers-cell">{code.currentViewers}</td>
                  <td className="total-cell">{code.totalViews}</td>
                  <td className="actions-cell">
                    {!code.isRevoked && (
                      <button
                        className="btn btn-danger btn-sm"
                        onClick={() =>
                          openRevokeModal(state, code.code, code.label)
                        }
                      >
                        Revoke
                      </button>
                    )}
                    <button className="btn btn-secondary btn-sm" disabled>
                      Analytics
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Revoke Confirmation Modal */}
      <Modal
        isOpen={state.revokeModal.isOpen}
        title="Revoke Access Code"
        onClose={() => closeRevokeModal(state)}
      >
        <div className="modal-body">
          <p>
            Are you sure you want to revoke this access code? All active viewers
            using this code will be disconnected immediately.
          </p>
          <div className="revoke-details">
            <div className="detail-row">
              <strong>Code:</strong> {state.revokeModal.code}
            </div>
            {state.revokeModal.label && (
              <div className="detail-row">
                <strong>Label:</strong> {state.revokeModal.label}
              </div>
            )}
          </div>

          {state.revokeModal.error && (
            <div className="error-message">{state.revokeModal.error}</div>
          )}

          <div className="modal-actions">
            <button
              className="btn btn-secondary"
              onClick={() => closeRevokeModal(state)}
              disabled={state.revokeModal.isSubmitting}
            >
              Cancel
            </button>
            <button
              className="btn btn-danger"
              onClick={() =>
                submitRevoke(
                  state,
                  props.studioId,
                  props.studioName,
                  props.rooms,
                )
              }
              disabled={state.revokeModal.isSubmitting}
            >
              {state.revokeModal.isSubmitting ? "Revoking..." : "Revoke Code"}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
