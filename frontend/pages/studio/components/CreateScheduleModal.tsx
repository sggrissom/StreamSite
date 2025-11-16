import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import "./CreateScheduleModal-styles";

type Room = {
  id: number;
  name: string;
};

type CreateScheduleModalProps = {
  studioId: number;
  rooms: Room[];
  editingSchedule: server.ClassSchedule | null;
  onClose: () => void;
  onSuccess: () => void;
};

type CreateScheduleModalState = {
  // Basic fields
  roomId: number;
  name: string;
  description: string;

  // Schedule type
  isRecurring: boolean;

  // One-time fields
  startDate: string; // YYYY-MM-DD
  startTime: string; // HH:MM
  endDate: string;
  endTime: string;

  // Recurring fields
  recurStartDate: string;
  recurEndDate: string;
  recurWeekdays: boolean[]; // [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
  recurTimeStart: string;
  recurTimeEnd: string;
  recurTimezone: string;

  // Camera automation
  preRollMinutes: number;
  postRollMinutes: number;
  autoStartCamera: boolean;
  autoStopCamera: boolean;

  // UI state
  isSubmitting: boolean;
  errors: { [key: string]: string };
};

const TIMEZONES = [
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "America/Phoenix",
  "America/Anchorage",
  "Pacific/Honolulu",
  "UTC",
  "Europe/London",
  "Europe/Paris",
  "Asia/Tokyo",
];

const WEEKDAY_NAMES = ["S", "M", "T", "W", "T", "F", "S"];
const WEEKDAY_FULL_NAMES = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

// Helper function to convert UTC ISO string to local date string (YYYY-MM-DD)
function toLocalDateString(isoString: string): string {
  const date = new Date(isoString);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

const useCreateScheduleModal = vlens.declareHook(
  (
    cacheKey: string,
    editingSchedule: server.ClassSchedule | null,
    defaultRoomId: number,
  ): CreateScheduleModalState => {
    const editing = editingSchedule;

    // Initialize from editing schedule if present
    const state: CreateScheduleModalState = {
      roomId: editing?.roomId || defaultRoomId,
      name: editing?.name || "",
      description: editing?.description || "",
      isRecurring: editing?.isRecurring || false,

      // One-time fields
      startDate:
        editing && !editing.isRecurring
          ? toLocalDateString(editing.startTime)
          : "",
      startTime:
        editing && !editing.isRecurring
          ? new Date(editing.startTime).toTimeString().slice(0, 5)
          : "",
      endDate:
        editing && !editing.isRecurring
          ? toLocalDateString(editing.endTime)
          : "",
      endTime:
        editing && !editing.isRecurring
          ? new Date(editing.endTime).toTimeString().slice(0, 5)
          : "",

      // Recurring fields
      recurStartDate:
        editing && editing.isRecurring
          ? toLocalDateString(editing.recurStartDate)
          : "",
      recurEndDate:
        editing && editing.isRecurring && editing.recurEndDate
          ? toLocalDateString(editing.recurEndDate)
          : "",
      recurWeekdays:
        editing && editing.isRecurring
          ? [0, 1, 2, 3, 4, 5, 6].map((i) => editing.recurWeekdays.includes(i))
          : [false, false, false, false, false, false, false],
      recurTimeStart: editing?.recurTimeStart || "",
      recurTimeEnd: editing?.recurTimeEnd || "",
      recurTimezone: editing?.recurTimezone || "America/New_York",

      // Camera automation
      preRollMinutes: editing?.preRollMinutes ?? 5,
      postRollMinutes: editing?.postRollMinutes ?? 2,
      autoStartCamera: editing?.autoStartCamera ?? true,
      autoStopCamera: editing?.autoStopCamera ?? true,

      isSubmitting: false,
      errors: {},
    };

    return state;
  },
);

function toggleWeekday(state: CreateScheduleModalState, index: number) {
  state.recurWeekdays[index] = !state.recurWeekdays[index];
  vlens.scheduleRedraw();
}

function handleRoomChange(state: CreateScheduleModalState, e: Event) {
  state.roomId = parseInt((e.target as HTMLSelectElement).value, 10);
  vlens.scheduleRedraw();
}

function handleNameInput(state: CreateScheduleModalState, e: Event) {
  state.name = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleDescriptionInput(state: CreateScheduleModalState, e: Event) {
  state.description = (e.target as HTMLTextAreaElement).value;
  vlens.scheduleRedraw();
}

function handleTypeChangeToOneTime(state: CreateScheduleModalState) {
  state.isRecurring = false;
  vlens.scheduleRedraw();
}

function handleTypeChangeToRecurring(state: CreateScheduleModalState) {
  state.isRecurring = true;
  vlens.scheduleRedraw();
}

function handleStartDateInput(state: CreateScheduleModalState, e: Event) {
  state.startDate = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleStartTimeInput(state: CreateScheduleModalState, e: Event) {
  state.startTime = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleEndDateInput(state: CreateScheduleModalState, e: Event) {
  state.endDate = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleEndTimeInput(state: CreateScheduleModalState, e: Event) {
  state.endTime = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleRecurTimeStartInput(state: CreateScheduleModalState, e: Event) {
  state.recurTimeStart = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleRecurTimeEndInput(state: CreateScheduleModalState, e: Event) {
  state.recurTimeEnd = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleTimezoneChange(state: CreateScheduleModalState, e: Event) {
  state.recurTimezone = (e.target as HTMLSelectElement).value;
  vlens.scheduleRedraw();
}

function handleRecurStartDateInput(state: CreateScheduleModalState, e: Event) {
  state.recurStartDate = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleRecurEndDateInput(state: CreateScheduleModalState, e: Event) {
  state.recurEndDate = (e.target as HTMLInputElement).value;
  vlens.scheduleRedraw();
}

function handleAutoStartChange(state: CreateScheduleModalState, e: Event) {
  state.autoStartCamera = (e.target as HTMLInputElement).checked;
  vlens.scheduleRedraw();
}

function handleAutoStopChange(state: CreateScheduleModalState, e: Event) {
  state.autoStopCamera = (e.target as HTMLInputElement).checked;
  vlens.scheduleRedraw();
}

function handlePreRollInput(state: CreateScheduleModalState, e: Event) {
  state.preRollMinutes =
    parseInt((e.target as HTMLInputElement).value, 10) || 0;
  vlens.scheduleRedraw();
}

function handlePostRollInput(state: CreateScheduleModalState, e: Event) {
  state.postRollMinutes =
    parseInt((e.target as HTMLInputElement).value, 10) || 0;
  vlens.scheduleRedraw();
}

function validateForm(state: CreateScheduleModalState): boolean {
  const errors: { [key: string]: string } = {};

  // Basic validation
  if (!state.name.trim()) {
    errors.name = "Name is required";
  }

  if (state.roomId === 0) {
    errors.roomId = "Please select a room";
  }

  if (!state.isRecurring) {
    // One-time validation
    if (!state.startDate) {
      errors.startDate = "Start date is required";
    }
    if (!state.startTime) {
      errors.startTime = "Start time is required";
    }
    if (!state.endDate) {
      errors.endDate = "End date is required";
    }
    if (!state.endTime) {
      errors.endTime = "End time is required";
    }

    // Check end after start
    if (state.startDate && state.endDate && state.startTime && state.endTime) {
      const start = new Date(`${state.startDate}T${state.startTime}`);
      const end = new Date(`${state.endDate}T${state.endTime}`);
      if (end <= start) {
        errors.endDate = "End time must be after start time";
      }
    }
  } else {
    // Recurring validation
    if (!state.recurStartDate) {
      errors.recurStartDate = "Start date is required";
    }
    if (!state.recurTimeStart) {
      errors.recurTimeStart = "Start time is required";
    }
    if (!state.recurTimeEnd) {
      errors.recurTimeEnd = "End time is required";
    }

    // Check at least one weekday selected
    if (!state.recurWeekdays.some((checked) => checked)) {
      errors.recurWeekdays = "Please select at least one day";
    }

    // Check end time after start time
    if (state.recurTimeStart && state.recurTimeEnd) {
      if (state.recurTimeEnd <= state.recurTimeStart) {
        errors.recurTimeEnd = "End time must be after start time";
      }
    }
  }

  // Validate pre/post-roll
  if (state.preRollMinutes < 0 || state.preRollMinutes > 60) {
    errors.preRollMinutes = "Pre-roll must be between 0 and 60 minutes";
  }
  if (state.postRollMinutes < 0 || state.postRollMinutes > 60) {
    errors.postRollMinutes = "Post-roll must be between 0 and 60 minutes";
  }

  state.errors = errors;
  return Object.keys(errors).length === 0;
}

async function handleSubmit(
  state: CreateScheduleModalState,
  props: CreateScheduleModalProps,
) {
  if (!validateForm(state)) {
    vlens.scheduleRedraw();
    return;
  }

  state.isSubmitting = true;
  state.errors = {};
  vlens.scheduleRedraw();

  try {
    if (props.editingSchedule) {
      // Update existing schedule - all fields start as null except required ones
      const request: server.UpdateClassScheduleRequest = {
        scheduleId: props.editingSchedule.id,
        name: state.name.trim(),
        description: state.description.trim(),
        preRollMinutes: state.preRollMinutes,
        postRollMinutes: state.postRollMinutes,
        autoStartCamera: state.autoStartCamera,
        autoStopCamera: state.autoStopCamera,
        startTime: null,
        endTime: null,
        recurStartDate: null,
        recurEndDate: null,
        recurWeekdays: [], // Must be array, not null
        recurTimeStart: null,
        recurTimeEnd: null,
        recurTimezone: null,
      };

      // Add schedule-specific fields
      if (!state.isRecurring) {
        request.startTime = new Date(
          `${state.startDate}T${state.startTime}`,
        ).toISOString();
        request.endTime = new Date(
          `${state.endDate}T${state.endTime}`,
        ).toISOString();
      } else {
        request.recurStartDate = new Date(state.recurStartDate).toISOString();
        if (state.recurEndDate) {
          request.recurEndDate = new Date(state.recurEndDate).toISOString();
        }
        request.recurWeekdays = state.recurWeekdays
          .map((checked, i) => (checked ? i : -1))
          .filter((i) => i !== -1);
        request.recurTimeStart = state.recurTimeStart;
        request.recurTimeEnd = state.recurTimeEnd;
        request.recurTimezone = state.recurTimezone;
      }

      const [resp, err] = await server.UpdateClassSchedule(request);

      if (err || !resp || !resp.success) {
        state.errors.submit = err || "Failed to update schedule";
        state.isSubmitting = false;
        vlens.scheduleRedraw();
        return;
      }
    } else {
      // Create new schedule - all fields required, use defaults for irrelevant ones
      const request: server.CreateClassScheduleRequest = {
        roomId: state.roomId,
        name: state.name.trim(),
        description: state.description.trim(),
        isRecurring: state.isRecurring,
        preRollMinutes: state.preRollMinutes,
        postRollMinutes: state.postRollMinutes,
        autoStartCamera: state.autoStartCamera,
        autoStopCamera: state.autoStopCamera,
        // Initialize all schedule fields with defaults
        startTime: "",
        endTime: "",
        recurStartDate: "",
        recurEndDate: "",
        recurWeekdays: [],
        recurTimeStart: "",
        recurTimeEnd: "",
        recurTimezone: "",
      };

      // Set schedule-specific fields based on type
      if (!state.isRecurring) {
        request.startTime = new Date(
          `${state.startDate}T${state.startTime}`,
        ).toISOString();
        request.endTime = new Date(
          `${state.endDate}T${state.endTime}`,
        ).toISOString();
      } else {
        request.recurStartDate = new Date(state.recurStartDate).toISOString();
        if (state.recurEndDate) {
          request.recurEndDate = new Date(state.recurEndDate).toISOString();
        }
        request.recurWeekdays = state.recurWeekdays
          .map((checked, i) => (checked ? i : -1))
          .filter((i) => i !== -1);
        request.recurTimeStart = state.recurTimeStart;
        request.recurTimeEnd = state.recurTimeEnd;
        request.recurTimezone = state.recurTimezone;
      }

      const [resp, err] = await server.CreateClassSchedule(request);

      if (err || !resp) {
        state.errors.submit = err || "Failed to create schedule";
        state.isSubmitting = false;
        vlens.scheduleRedraw();
        return;
      }
    }

    // Success
    props.onSuccess();
    props.onClose();
  } catch (error) {
    state.errors.submit = "An unexpected error occurred";
    state.isSubmitting = false;
    vlens.scheduleRedraw();
  }
}

export function CreateScheduleModal(props: CreateScheduleModalProps) {
  // Use different cache keys for create vs edit to reset state properly
  const cacheKey = props.editingSchedule
    ? `edit-${props.editingSchedule.id}`
    : "create";

  const state = useCreateScheduleModal(
    cacheKey,
    props.editingSchedule,
    props.rooms[0]?.id ?? 0,
  );

  return (
    <Modal
      isOpen={true}
      title={props.editingSchedule ? "Edit Schedule" : "Create Schedule"}
      onClose={props.onClose}
    >
      <div className="create-schedule-modal">
        {/* Room Selection */}
        <div className="schedule-form-group">
          <label className="schedule-form-label schedule-form-label-required">
            Room
          </label>
          <select
            className="schedule-form-select"
            value={state.roomId}
            onChange={vlens.cachePartial(handleRoomChange, state)}
            disabled={!!props.editingSchedule} // Can't change room when editing
          >
            {props.rooms.map((room) => (
              <option key={room.id} value={room.id}>
                {room.name}
              </option>
            ))}
          </select>
          {state.errors.roomId && (
            <div className="schedule-form-error">{state.errors.roomId}</div>
          )}
        </div>

        {/* Name */}
        <div className="schedule-form-group">
          <label className="schedule-form-label schedule-form-label-required">
            Class Name
          </label>
          <input
            type="text"
            className="schedule-form-input"
            value={state.name}
            onInput={vlens.cachePartial(handleNameInput, state)}
            placeholder="e.g., Math 101"
          />
          {state.errors.name && (
            <div className="schedule-form-error">{state.errors.name}</div>
          )}
        </div>

        {/* Description */}
        <div className="schedule-form-group">
          <label className="schedule-form-label">Description</label>
          <textarea
            className="schedule-form-input schedule-form-textarea"
            value={state.description}
            onInput={vlens.cachePartial(handleDescriptionInput, state)}
            placeholder="Optional description of the class"
          />
        </div>

        {/* Schedule Type Toggle */}
        <div className="schedule-form-group">
          <label className="schedule-form-label">Schedule Type</label>
          <div className="schedule-type-toggle">
            <div className="schedule-type-option">
              <input
                type="radio"
                id="type-onetime"
                name="scheduleType"
                className="schedule-type-radio"
                checked={!state.isRecurring}
                onChange={vlens.cachePartial(handleTypeChangeToOneTime, state)}
                disabled={!!props.editingSchedule} // Can't change type when editing
              />
              <label htmlFor="type-onetime" className="schedule-type-label">
                One-Time
              </label>
            </div>
            <div className="schedule-type-option">
              <input
                type="radio"
                id="type-recurring"
                name="scheduleType"
                className="schedule-type-radio"
                checked={state.isRecurring}
                onChange={vlens.cachePartial(
                  handleTypeChangeToRecurring,
                  state,
                )}
                disabled={!!props.editingSchedule} // Can't change type when editing
              />
              <label htmlFor="type-recurring" className="schedule-type-label">
                Recurring
              </label>
            </div>
          </div>
        </div>

        {/* One-Time Schedule Fields */}
        {!state.isRecurring && (
          <>
            <div className="schedule-form-group">
              <label className="schedule-form-label schedule-form-label-required">
                Start Date & Time
              </label>
              <div className="time-inputs-row">
                <div className="time-input-group">
                  <input
                    type="date"
                    className="schedule-form-input"
                    value={state.startDate}
                    onInput={vlens.cachePartial(handleStartDateInput, state)}
                  />
                  {state.errors.startDate && (
                    <div className="schedule-form-error">
                      {state.errors.startDate}
                    </div>
                  )}
                </div>
                <div className="time-input-group">
                  <input
                    type="time"
                    className="schedule-form-input"
                    value={state.startTime}
                    onInput={vlens.cachePartial(handleStartTimeInput, state)}
                  />
                  {state.errors.startTime && (
                    <div className="schedule-form-error">
                      {state.errors.startTime}
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="schedule-form-group">
              <label className="schedule-form-label schedule-form-label-required">
                End Date & Time
              </label>
              <div className="time-inputs-row">
                <div className="time-input-group">
                  <input
                    type="date"
                    className="schedule-form-input"
                    value={state.endDate}
                    onInput={vlens.cachePartial(handleEndDateInput, state)}
                  />
                  {state.errors.endDate && (
                    <div className="schedule-form-error">
                      {state.errors.endDate}
                    </div>
                  )}
                </div>
                <div className="time-input-group">
                  <input
                    type="time"
                    className="schedule-form-input"
                    value={state.endTime}
                    onInput={vlens.cachePartial(handleEndTimeInput, state)}
                  />
                  {state.errors.endTime && (
                    <div className="schedule-form-error">
                      {state.errors.endTime}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </>
        )}

        {/* Recurring Schedule Fields */}
        {state.isRecurring && (
          <>
            <div className="schedule-form-group">
              <label className="schedule-form-label schedule-form-label-required">
                Days of Week
              </label>
              <div className="weekday-selector">
                {WEEKDAY_NAMES.map((day, index) => (
                  <button
                    key={index}
                    type="button"
                    className={`weekday-button ${state.recurWeekdays[index] ? "selected" : ""}`}
                    onClick={() => toggleWeekday(state, index)}
                    title={WEEKDAY_FULL_NAMES[index]}
                  >
                    {day}
                  </button>
                ))}
              </div>
              {state.errors.recurWeekdays && (
                <div className="schedule-form-error">
                  {state.errors.recurWeekdays}
                </div>
              )}
            </div>

            <div className="schedule-form-group">
              <label className="schedule-form-label schedule-form-label-required">
                Time
              </label>
              <div className="time-inputs-row">
                <div className="time-input-group">
                  <input
                    type="time"
                    className="schedule-form-input"
                    value={state.recurTimeStart}
                    onInput={vlens.cachePartial(
                      handleRecurTimeStartInput,
                      state,
                    )}
                  />
                  {state.errors.recurTimeStart && (
                    <div className="schedule-form-error">
                      {state.errors.recurTimeStart}
                    </div>
                  )}
                </div>
                <span className="time-separator">to</span>
                <div className="time-input-group">
                  <input
                    type="time"
                    className="schedule-form-input"
                    value={state.recurTimeEnd}
                    onInput={vlens.cachePartial(handleRecurTimeEndInput, state)}
                  />
                  {state.errors.recurTimeEnd && (
                    <div className="schedule-form-error">
                      {state.errors.recurTimeEnd}
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="schedule-form-group">
              <label className="schedule-form-label schedule-form-label-required">
                Timezone
              </label>
              <select
                className="schedule-form-select"
                value={state.recurTimezone}
                onChange={vlens.cachePartial(handleTimezoneChange, state)}
              >
                {TIMEZONES.map((tz) => (
                  <option key={tz} value={tz}>
                    {tz}
                  </option>
                ))}
              </select>
            </div>

            <div className="schedule-form-group">
              <label className="schedule-form-label schedule-form-label-required">
                Start Date
              </label>
              <input
                type="date"
                className="schedule-form-input"
                value={state.recurStartDate}
                onInput={vlens.cachePartial(handleRecurStartDateInput, state)}
              />
              {state.errors.recurStartDate && (
                <div className="schedule-form-error">
                  {state.errors.recurStartDate}
                </div>
              )}
              <div className="form-helper-text">
                When should this recurring schedule start?
              </div>
            </div>

            <div className="schedule-form-group">
              <label className="schedule-form-label">End Date (Optional)</label>
              <input
                type="date"
                className="schedule-form-input"
                value={state.recurEndDate}
                onInput={vlens.cachePartial(handleRecurEndDateInput, state)}
              />
              <div className="form-helper-text">
                Leave blank for indefinite recurrence
              </div>
            </div>
          </>
        )}

        {/* Camera Automation */}
        <div className="camera-automation-section">
          <div className="camera-automation-title">Camera Automation</div>

          <div className="checkbox-toggle">
            <input
              type="checkbox"
              id="auto-start"
              checked={state.autoStartCamera}
              onChange={vlens.cachePartial(handleAutoStartChange, state)}
            />
            <label htmlFor="auto-start">Auto-start camera</label>
          </div>

          <div className="checkbox-toggle">
            <input
              type="checkbox"
              id="auto-stop"
              checked={state.autoStopCamera}
              onChange={vlens.cachePartial(handleAutoStopChange, state)}
            />
            <label htmlFor="auto-stop">Auto-stop camera</label>
          </div>

          <div className="schedule-form-group" style={{ marginTop: "0.75rem" }}>
            <label className="schedule-form-label">Pre-roll (minutes)</label>
            <div className="number-input-group">
              <input
                type="number"
                className="schedule-form-input number-input-small"
                value={state.preRollMinutes}
                onInput={vlens.cachePartial(handlePreRollInput, state)}
                min={0}
                max={60}
              />
              <span className="number-input-suffix">
                minutes before class starts
              </span>
            </div>
            {state.errors.preRollMinutes && (
              <div className="schedule-form-error">
                {state.errors.preRollMinutes}
              </div>
            )}
          </div>

          <div className="schedule-form-group">
            <label className="schedule-form-label">Post-roll (minutes)</label>
            <div className="number-input-group">
              <input
                type="number"
                className="schedule-form-input number-input-small"
                value={state.postRollMinutes}
                onInput={vlens.cachePartial(handlePostRollInput, state)}
                min={0}
                max={60}
              />
              <span className="number-input-suffix">
                minutes after class ends
              </span>
            </div>
            {state.errors.postRollMinutes && (
              <div className="schedule-form-error">
                {state.errors.postRollMinutes}
              </div>
            )}
          </div>
        </div>

        {/* Submit Error */}
        {state.errors.submit && (
          <div className="schedule-form-error" style={{ marginTop: "1rem" }}>
            {state.errors.submit}
          </div>
        )}

        {/* Actions */}
        <div className="schedule-form-actions">
          <button
            type="button"
            className="btn-schedule-cancel"
            onClick={props.onClose}
            disabled={state.isSubmitting}
          >
            Cancel
          </button>
          <button
            type="button"
            className="btn-schedule-submit"
            onClick={() => handleSubmit(state, props)}
            disabled={state.isSubmitting}
          >
            {state.isSubmitting
              ? "Saving..."
              : props.editingSchedule
                ? "Update Schedule"
                : "Create Schedule"}
          </button>
        </div>
      </div>
    </Modal>
  );
}
