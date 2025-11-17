import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import "./ExecutionLogsModal-styles";

type ExecutionLogsModalProps = {
  isOpen: boolean;
  onClose: () => void;
  scheduleId?: number;
  roomId?: number;
  title: string;
};

type ModalState = {
  logs: server.ScheduleExecutionLog[];
  isLoading: boolean;
  error: string;
  total: number;
  limit: number;
  offset: number;
};

const useModalState = vlens.declareHook((id: string): ModalState => {
  const state: ModalState = {
    logs: [],
    isLoading: true,
    error: "",
    total: 0,
    limit: 50,
    offset: 0,
  };

  return state;
});

async function loadLogs(
  state: ModalState,
  scheduleId?: number,
  roomId?: number,
) {
  state.isLoading = true;
  state.error = "";
  vlens.scheduleRedraw();

  const req: server.GetScheduleExecutionLogsRequest = {
    scheduleId: scheduleId || null,
    roomId: roomId || null,
    limit: state.limit,
    offset: state.offset,
  };

  const [resp, err] = await server.GetScheduleExecutionLogs(req);

  state.isLoading = false;

  if (err || !resp) {
    state.error = err || "Failed to load execution logs";
    vlens.scheduleRedraw();
    return;
  }

  state.logs = resp.logs || [];
  state.total = resp.total || 0;
  vlens.scheduleRedraw();
}

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function getActionLabel(action: string): string {
  switch (action) {
    case "start_camera":
      return "Start Camera";
    case "stop_camera":
      return "Stop Camera";
    case "skip_already_running":
      return "Skip (Already Running)";
    default:
      return action;
  }
}

function getActionIcon(action: string): string {
  switch (action) {
    case "start_camera":
      return "▶️";
    case "stop_camera":
      return "⏹️";
    case "skip_already_running":
      return "⏭️";
    default:
      return "•";
  }
}

function handlePrevPage(
  state: ModalState,
  scheduleId?: number,
  roomId?: number,
) {
  if (state.offset > 0) {
    state.offset = Math.max(0, state.offset - state.limit);
    loadLogs(state, scheduleId, roomId);
  }
}

function handleNextPage(
  state: ModalState,
  scheduleId?: number,
  roomId?: number,
) {
  if (state.offset + state.limit < state.total) {
    state.offset += state.limit;
    loadLogs(state, scheduleId, roomId);
  }
}

export function ExecutionLogsModal(props: ExecutionLogsModalProps) {
  const modalId = `logs-${props.scheduleId || 0}-${props.roomId || 0}`;
  const state = useModalState(modalId);

  // Load logs when modal opens
  if (props.isOpen && state.isLoading && state.logs.length === 0) {
    loadLogs(state, props.scheduleId, props.roomId);
  }

  if (!props.isOpen) {
    return null;
  }

  const currentPage = Math.floor(state.offset / state.limit) + 1;
  const totalPages = Math.ceil(state.total / state.limit);

  return (
    <Modal isOpen={props.isOpen} onClose={props.onClose} title={props.title}>
      <div className="execution-logs-modal">
        {state.error && <div className="error-message">{state.error}</div>}

        {state.isLoading ? (
          <div className="loading-message">Loading execution logs...</div>
        ) : state.logs.length === 0 ? (
          <div className="empty-message">
            No execution logs found. The scheduler will create logs when it
            automatically starts or stops cameras for this schedule.
          </div>
        ) : (
          <>
            <div className="logs-header">
              <div className="logs-count">
                Showing {state.offset + 1}-
                {Math.min(state.offset + state.limit, state.total)} of{" "}
                {state.total} logs
              </div>
            </div>

            <div className="logs-table">
              <div className="logs-table-header">
                <div className="col-timestamp">Timestamp</div>
                <div className="col-action">Action</div>
                <div className="col-status">Status</div>
                <div className="col-error">Details</div>
              </div>

              <div className="logs-table-body">
                {state.logs.map((log) => (
                  <div
                    key={log.id}
                    className={`log-row ${log.success ? "success" : "error"}`}
                  >
                    <div className="col-timestamp">
                      {formatTimestamp(log.timestamp)}
                    </div>
                    <div className="col-action">
                      <span className="action-icon">
                        {getActionIcon(log.action)}
                      </span>
                      <span className="action-label">
                        {getActionLabel(log.action)}
                      </span>
                    </div>
                    <div className="col-status">
                      {log.success ? (
                        <span className="status-badge success">✓ Success</span>
                      ) : (
                        <span className="status-badge error">✗ Error</span>
                      )}
                    </div>
                    <div className="col-error">{log.errorMsg || "-"}</div>
                  </div>
                ))}
              </div>
            </div>

            {totalPages > 1 && (
              <div className="pagination">
                <button
                  onClick={() =>
                    handlePrevPage(state, props.scheduleId, props.roomId)
                  }
                  disabled={state.offset === 0}
                  className="pagination-btn"
                >
                  ← Previous
                </button>
                <span className="pagination-info">
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  onClick={() =>
                    handleNextPage(state, props.scheduleId, props.roomId)
                  }
                  disabled={state.offset + state.limit >= state.total}
                  className="pagination-btn"
                >
                  Next →
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  );
}
