import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import { CreateScheduleModal } from "./CreateScheduleModal";
import { ClassPermissionsModal } from "./ClassPermissionsModal";
import "./SchedulesSection-styles";

type Studio = {
  id: number;
  name: string;
};

type Room = {
  id: number;
  name: string;
  roomNumber: number;
};

type SchedulesSectionProps = {
  studio: Studio;
  rooms: Room[];
  canManageRooms: boolean;
};

type SchedulesSectionState = {
  schedules: server.ClassSchedule[];
  isLoading: boolean;
  error: string;
  filterRoomId: number; // 0 = all rooms
  showCreateModal: boolean;
  editingSchedule: server.ClassSchedule | null;
  deleteConfirmSchedule: server.ClassSchedule | null;
  permissionsSchedule: server.ClassSchedule | null;
  isDeleting: boolean;
};

const useSchedulesSection = vlens.declareHook(
  (studioId: number): SchedulesSectionState => {
    const state: SchedulesSectionState = {
      schedules: [],
      isLoading: true,
      error: "",
      filterRoomId: 0,
      showCreateModal: false,
      editingSchedule: null,
      deleteConfirmSchedule: null,
      permissionsSchedule: null,
      isDeleting: false,
    };

    loadSchedules(state, studioId);

    return state;
  },
);

async function loadSchedules(state: SchedulesSectionState, studioId: number) {
  state.isLoading = true;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.ListClassSchedules({
    studioId,
    roomId: null,
  });

  state.isLoading = false;

  if (err || !resp) {
    state.error = err || "Failed to load schedules";
    vlens.scheduleRedraw();
    return;
  }

  state.schedules = resp.schedules || [];
  vlens.scheduleRedraw();
}

function openCreateModal(state: SchedulesSectionState) {
  state.showCreateModal = true;
  state.editingSchedule = null;
  vlens.scheduleRedraw();
}

function closeCreateModal(state: SchedulesSectionState) {
  state.showCreateModal = false;
  state.editingSchedule = null;
  vlens.scheduleRedraw();
}

function openEditModal(
  state: SchedulesSectionState,
  schedule: server.ClassSchedule,
) {
  state.editingSchedule = schedule;
  state.showCreateModal = true;
  vlens.scheduleRedraw();
}

function openDeleteConfirm(
  state: SchedulesSectionState,
  schedule: server.ClassSchedule,
) {
  state.deleteConfirmSchedule = schedule;
  vlens.scheduleRedraw();
}

function closeDeleteConfirm(state: SchedulesSectionState) {
  state.deleteConfirmSchedule = null;
  vlens.scheduleRedraw();
}

function openPermissionsModal(
  state: SchedulesSectionState,
  schedule: server.ClassSchedule,
) {
  state.permissionsSchedule = schedule;
  vlens.scheduleRedraw();
}

function closePermissionsModal(state: SchedulesSectionState) {
  state.permissionsSchedule = null;
  vlens.scheduleRedraw();
}

async function confirmDelete(
  state: SchedulesSectionState,
  scheduleId: number,
  studioId: number,
) {
  state.isDeleting = true;
  vlens.scheduleRedraw();

  const [resp, err] = await server.DeleteClassSchedule({ scheduleId });

  if (err || !resp || !resp.success) {
    alert(err || "Failed to delete schedule");
    state.isDeleting = false;
    vlens.scheduleRedraw();
    return;
  }

  // Reload schedules
  state.deleteConfirmSchedule = null;
  state.isDeleting = false;
  await loadSchedules(state, studioId);
}

function getScheduleStatus(schedule: server.ClassSchedule): {
  label: string;
  className: string;
  icon: string;
} {
  const now = new Date();

  // For one-time schedules
  if (!schedule.isRecurring) {
    const startTime = new Date(schedule.startTime);
    const endTime = new Date(schedule.endTime);

    if (now >= startTime && now <= endTime) {
      return { label: "Live", className: "live", icon: "🟢" };
    }

    // Check if upcoming (within next hour)
    const oneHour = 60 * 60 * 1000;
    if (startTime.getTime() - now.getTime() <= oneHour && now < startTime) {
      const minutesUntil = Math.floor(
        (startTime.getTime() - now.getTime()) / (60 * 1000),
      );
      return {
        label: `Soon (${minutesUntil}min)`,
        className: "upcoming",
        icon: "⏰",
      };
    }

    if (now > endTime) {
      return { label: "Past", className: "idle", icon: "" };
    }

    return { label: "Upcoming", className: "idle", icon: "" };
  }

  // For recurring schedules - check if currently live
  try {
    // Get current time components in the schedule's timezone using proper API
    const formatter = new Intl.DateTimeFormat("en-US", {
      timeZone: schedule.recurTimezone,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });

    const parts = formatter.formatToParts(now);
    const getPart = (type: string) =>
      parseInt(parts.find((p) => p.type === type)?.value || "0");

    const year = getPart("year");
    const month = getPart("month");
    const day = getPart("day");
    const hour = getPart("hour");
    const minute = getPart("minute");
    const second = getPart("second");

    // Create a date representing "now" in the schedule's timezone
    // Note: month is 0-indexed in Date constructor
    const nowInTz = new Date(year, month - 1, day, hour, minute, second);
    const weekday = nowInTz.getDay(); // 0=Sun, 6=Sat

    // Check if today is a scheduled weekday
    const isScheduledDay = schedule.recurWeekdays.includes(weekday);
    if (!isScheduledDay) {
      return { label: "Recurring", className: "idle", icon: "🔄" };
    }

    // Check if we're within the recurrence date range (date-only comparison)
    // Format current date in schedule's timezone as YYYY-MM-DD string
    const nowDateStr = `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;

    if (schedule.recurStartDate) {
      // Extract date-only from ISO string (e.g., "2025-11-15T00:00:00Z" -> "2025-11-15")
      const startDateStr = schedule.recurStartDate.split("T")[0];

      // String comparison works correctly for ISO date format (YYYY-MM-DD)
      if (nowDateStr < startDateStr) {
        return { label: "Recurring", className: "idle", icon: "🔄" };
      }
    }
    if (schedule.recurEndDate) {
      // Extract date-only from ISO string
      const endDateStr = schedule.recurEndDate.split("T")[0];

      // String comparison works correctly for ISO date format (YYYY-MM-DD)
      if (nowDateStr > endDateStr) {
        return { label: "Past", className: "idle", icon: "" };
      }
    }

    // Parse the class time range (HH:MM format)
    const [startHour, startMin] = schedule.recurTimeStart
      .split(":")
      .map(Number);
    const [endHour, endMin] = schedule.recurTimeEnd.split(":").map(Number);

    // Build today's class time window in the schedule's timezone
    const classStart = new Date(nowInTz);
    classStart.setHours(startHour, startMin, 0, 0);

    const classEnd = new Date(nowInTz);
    classEnd.setHours(endHour, endMin, 0, 0);

    // Check if currently live
    if (nowInTz >= classStart && nowInTz <= classEnd) {
      return { label: "Live Now", className: "live", icon: "🟢" };
    }

    // Check if upcoming today (within next hour)
    const oneHour = 60 * 60 * 1000;
    if (
      classStart.getTime() - nowInTz.getTime() <= oneHour &&
      nowInTz < classStart
    ) {
      const minutesUntil = Math.floor(
        (classStart.getTime() - nowInTz.getTime()) / (60 * 1000),
      );
      return {
        label: `Soon (${minutesUntil}min)`,
        className: "upcoming",
        icon: "⏰",
      };
    }

    return { label: "Recurring", className: "idle", icon: "🔄" };
  } catch (error) {
    // If timezone parsing fails, fall back to simple recurring label
    console.error("Error checking recurring schedule status:", error);
    return { label: "Recurring", className: "idle", icon: "🔄" };
  }
}

function formatSchedulePattern(schedule: server.ClassSchedule): string {
  if (!schedule.isRecurring) {
    const startDate = new Date(schedule.startTime);
    return `${startDate.toLocaleDateString()} ${startDate.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
  }

  const weekdayNames = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
  const days = schedule.recurWeekdays
    .map((d: number) => weekdayNames[d])
    .join(", ");
  return `${days} ${schedule.recurTimeStart}-${schedule.recurTimeEnd}`;
}

function formatScheduleType(schedule: server.ClassSchedule): string {
  return schedule.isRecurring ? "Recurring" : "One-time";
}

export function SchedulesSection(props: SchedulesSectionProps) {
  const state = useSchedulesSection(props.studio.id);

  // Filter schedules by room
  const filteredSchedules = state.schedules.filter((schedule) => {
    if (state.filterRoomId === 0) return true;
    return schedule.roomId === state.filterRoomId;
  });

  // Get room name by ID
  const getRoomName = (roomId: number): string => {
    const room = props.rooms.find((r) => r.id === roomId);
    return room ? room.name : `Room ${roomId}`;
  };

  return (
    <div className="schedules-section">
      <div className="schedules-header">
        <h2 className="schedules-title">Class Schedules</h2>
        {props.canManageRooms && (
          <button
            className="btn btn-primary"
            onClick={() => openCreateModal(state)}
          >
            + Create Schedule
          </button>
        )}
      </div>

      {/* Filters */}
      <div className="schedules-filters">
        <span className="schedules-filter-label">Filter by room:</span>
        <select
          className="schedules-filter-select"
          value={state.filterRoomId}
          onChange={(e) => {
            state.filterRoomId = parseInt(
              (e.target as HTMLSelectElement).value,
              10,
            );
            vlens.scheduleRedraw();
          }}
        >
          <option value={0}>All Rooms</option>
          {props.rooms.map((room) => (
            <option key={room.id} value={room.id}>
              {room.name}
            </option>
          ))}
        </select>
      </div>

      {/* Loading state */}
      {state.isLoading && (
        <div className="schedules-loading">Loading schedules...</div>
      )}

      {/* Error state */}
      {state.error && (
        <div className="error-message" style={{ padding: "1rem" }}>
          {state.error}
        </div>
      )}

      {/* Empty state */}
      {!state.isLoading && !state.error && filteredSchedules.length === 0 && (
        <div className="schedules-empty">
          <div className="schedules-empty-icon">📅</div>
          <div className="schedules-empty-text">
            {state.filterRoomId === 0
              ? "No class schedules yet"
              : "No schedules for this room"}
          </div>
          {props.canManageRooms && (
            <button
              className="btn btn-primary"
              onClick={() => openCreateModal(state)}
              style={{ marginTop: "1rem" }}
            >
              Create First Schedule
            </button>
          )}
        </div>
      )}

      {/* Schedules table */}
      {!state.isLoading && !state.error && filteredSchedules.length > 0 && (
        <table className="schedules-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Room</th>
              <th>Schedule</th>
              <th>Status</th>
              {props.canManageRooms && <th>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {filteredSchedules.map((schedule) => {
              const status = getScheduleStatus(schedule);
              return (
                <tr key={schedule.id}>
                  <td>
                    <div className="schedule-name">{schedule.name}</div>
                    {schedule.description && (
                      <div className="schedule-description">
                        {schedule.description}
                      </div>
                    )}
                  </td>
                  <td>{getRoomName(schedule.roomId)}</td>
                  <td>
                    <div className="schedule-pattern">
                      {formatSchedulePattern(schedule)}
                    </div>
                    <div className="schedule-time">
                      {formatScheduleType(schedule)}
                      {schedule.recurTimezone && ` (${schedule.recurTimezone})`}
                    </div>
                  </td>
                  <td>
                    <span className={`schedule-status ${status.className}`}>
                      {status.icon && <span>{status.icon}</span>}
                      <span>{status.label}</span>
                    </span>
                  </td>
                  {props.canManageRooms && (
                    <td>
                      <div className="schedule-actions">
                        <button
                          className="btn-schedule"
                          onClick={() => openPermissionsModal(state, schedule)}
                        >
                          Permissions
                        </button>
                        <button
                          className="btn-schedule"
                          onClick={() => openEditModal(state, schedule)}
                        >
                          Edit
                        </button>
                        <button
                          className="btn-schedule-delete"
                          onClick={() => openDeleteConfirm(state, schedule)}
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  )}
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      {/* Create/Edit Modal */}
      {state.showCreateModal && (
        <CreateScheduleModal
          studioId={props.studio.id}
          rooms={props.rooms}
          editingSchedule={state.editingSchedule}
          onClose={() => closeCreateModal(state)}
          onSuccess={() => loadSchedules(state, props.studio.id)}
        />
      )}

      {/* Delete Confirmation Modal */}
      {state.deleteConfirmSchedule && (
        <Modal
          isOpen={true}
          title="Delete Schedule"
          onClose={() => closeDeleteConfirm(state)}
        >
          <div style={{ padding: "1rem" }}>
            <p>
              Are you sure you want to delete the schedule "
              {state.deleteConfirmSchedule.name}"?
            </p>
            <p style={{ color: "#666", fontSize: "0.875rem" }}>
              This action cannot be undone.
            </p>
            <div
              style={{
                display: "flex",
                gap: "0.5rem",
                marginTop: "1.5rem",
                justifyContent: "flex-end",
              }}
            >
              <button
                className="btn"
                onClick={() => closeDeleteConfirm(state)}
                disabled={state.isDeleting}
              >
                Cancel
              </button>
              <button
                className="btn btn-danger"
                onClick={() =>
                  confirmDelete(
                    state,
                    state.deleteConfirmSchedule!.id,
                    props.studio.id,
                  )
                }
                disabled={state.isDeleting}
              >
                {state.isDeleting ? "Deleting..." : "Delete Schedule"}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Class Permissions Modal */}
      {state.permissionsSchedule && (
        <ClassPermissionsModal
          isOpen={true}
          onClose={() => closePermissionsModal(state)}
          schedule={state.permissionsSchedule}
          studioId={props.studio.id}
        />
      )}
    </div>
  );
}
