import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import qrcode from "qrcode-generator";

type GenerateAccessCodeModalProps = {
  isOpen: boolean;
  onClose: () => void;
  codeType: number; // 0 = room, 1 = studio
  targetId: number;
  targetName: string;
  targetLabel: string; // "Room" or "Studio"
};

type ModalState = {
  isSubmitting: boolean;
  error: string;
  duration: string;
  maxViewers: number;
  label: string;
  generatedCode: string;
  shareUrl: string;
  qrDataUrl: string;
  showSuccess: boolean;
  copyCodeSuccess: boolean;
  copyUrlSuccess: boolean;
};

const useModalState = vlens.declareHook(
  (propsId: string): ModalState => ({
    isSubmitting: false,
    error: "",
    duration: "24h",
    maxViewers: 30,
    label: "",
    generatedCode: "",
    shareUrl: "",
    qrDataUrl: "",
    showSuccess: false,
    copyCodeSuccess: false,
    copyUrlSuccess: false,
  }),
);

function resetModal(state: ModalState) {
  state.error = "";
  state.duration = "24h";
  state.maxViewers = 30;
  state.label = "";
  state.generatedCode = "";
  state.shareUrl = "";
  state.qrDataUrl = "";
  state.showSuccess = false;
  state.copyCodeSuccess = false;
  state.copyUrlSuccess = false;
  state.isSubmitting = false;
  vlens.scheduleRedraw();
}

function durationToMinutes(duration: string): number {
  switch (duration) {
    case "1h":
      return 60;
    case "24h":
      return 1440;
    case "7d":
      return 10080;
    case "30d":
      return 43200;
    case "never":
      return -1; // Special value for "never expires"
    default:
      return 1440; // default to 24h
  }
}

// Convert GIF data URL to PNG data URL
function convertGifToPng(gifDataUrl: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      const canvas = document.createElement("canvas");
      canvas.width = img.width;
      canvas.height = img.height;
      const ctx = canvas.getContext("2d");
      if (!ctx) {
        reject(new Error("Failed to get canvas context"));
        return;
      }
      ctx.drawImage(img, 0, 0);
      resolve(canvas.toDataURL("image/png"));
    };
    img.onerror = () => reject(new Error("Failed to load image"));
    img.src = gifDataUrl;
  });
}

async function submitGenerateCode(
  state: ModalState,
  codeType: number,
  targetId: number,
) {
  if (state.maxViewers < 1) {
    state.error = "Max viewers must be at least 1";
    vlens.scheduleRedraw();
    return;
  }

  state.isSubmitting = true;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.GenerateAccessCode({
    type: codeType,
    targetId: targetId,
    durationMinutes: durationToMinutes(state.duration),
    maxViewers: state.maxViewers,
    label: state.label.trim() || "",
  });

  state.isSubmitting = false;

  if (err || !resp) {
    state.error = err || "Failed to generate access code";
    vlens.scheduleRedraw();
    return;
  }

  // Show success view with generated code
  state.generatedCode = resp.code || "";
  const hostname =
    window.location.hostname === "localhost"
      ? "localhost:3000"
      : "stream.grissom.zone";
  state.shareUrl = `http://${hostname}/watch/${resp.code}`;

  // Generate QR code for the share URL
  try {
    const qr = qrcode(0, "M");
    qr.addData(state.shareUrl);
    qr.make();
    const gifDataUrl = qr.createDataURL(4);
    // Convert GIF to PNG for proper download
    state.qrDataUrl = await convertGifToPng(gifDataUrl);
  } catch (qrError) {
    console.error("Failed to generate QR code:", qrError);
    state.qrDataUrl = "";
  }

  state.showSuccess = true;
  vlens.scheduleRedraw();
}

async function copyCode(state: ModalState) {
  try {
    await navigator.clipboard.writeText(state.generatedCode);
    state.copyCodeSuccess = true;
    vlens.scheduleRedraw();
    setTimeout(() => {
      state.copyCodeSuccess = false;
      vlens.scheduleRedraw();
    }, 2000);
  } catch (err) {
    state.error = "Failed to copy to clipboard";
    vlens.scheduleRedraw();
  }
}

async function copyShareUrl(state: ModalState) {
  try {
    await navigator.clipboard.writeText(state.shareUrl);
    state.copyUrlSuccess = true;
    vlens.scheduleRedraw();
    setTimeout(() => {
      state.copyUrlSuccess = false;
      vlens.scheduleRedraw();
    }, 2000);
  } catch (err) {
    state.error = "Failed to copy to clipboard";
    vlens.scheduleRedraw();
  }
}

function downloadQRCode(state: ModalState) {
  if (!state.qrDataUrl) return;

  const link = document.createElement("a");
  link.href = state.qrDataUrl;
  link.download = `access-code-${state.generatedCode}.png`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

function handleClose(state: ModalState, onClose: () => void) {
  resetModal(state);
  onClose();
}

export function GenerateAccessCodeModal(
  props: GenerateAccessCodeModalProps,
): preact.ComponentChild {
  const state = useModalState(`${props.codeType}-${props.targetId}`);

  // Reset state when modal is closed
  if (!props.isOpen && state.showSuccess) {
    resetModal(state);
  }

  const scopeDescription =
    props.codeType === 1
      ? `access to all rooms in this ${props.targetLabel.toLowerCase()}`
      : `access to this ${props.targetLabel.toLowerCase()}`;

  return (
    <Modal
      isOpen={props.isOpen}
      title={
        state.showSuccess
          ? `${props.targetLabel} Access Code Generated`
          : `Generate ${props.targetLabel} Access Code`
      }
      onClose={() => handleClose(state, props.onClose)}
      error={state.error}
      footer={
        state.showSuccess ? (
          <button
            className="btn btn-primary"
            onClick={() => handleClose(state, props.onClose)}
          >
            Done
          </button>
        ) : (
          <>
            <button
              className="btn btn-secondary"
              onClick={() => handleClose(state, props.onClose)}
              disabled={state.isSubmitting}
            >
              Cancel
            </button>
            <button
              className="btn btn-primary"
              onClick={() =>
                submitGenerateCode(state, props.codeType, props.targetId)
              }
              disabled={state.isSubmitting}
            >
              {state.isSubmitting ? "Generating..." : "Generate Code"}
            </button>
          </>
        )
      }
    >
      {state.showSuccess ? (
        // Success view - show generated code and share URL
        <div>
          <div className="form-group">
            <label>{props.targetLabel}</label>
            <div className="stream-key-room-name">{props.targetName}</div>
          </div>

          <div className="form-group">
            <label>Access Code</label>
            <div
              className="stream-key-display"
              style="font-size: 2rem; letter-spacing: 0.5rem; text-align: center; padding: 1.5rem;"
            >
              {state.generatedCode}
            </div>
            <small className="form-help">
              Share this 5-digit code with viewers to grant them{" "}
              {scopeDescription}
            </small>
          </div>

          <div className="form-group">
            <label>Share URL</label>
            <div className="stream-key-display">{state.shareUrl}</div>
            <small className="form-help">
              Direct link that viewers can use to access the{" "}
              {props.targetLabel.toLowerCase()}
            </small>
          </div>

          {state.qrDataUrl && (
            <div className="form-group">
              <label>QR Code</label>
              <div style="text-align: center; padding: 1rem; background: white; border: 1px solid var(--border); border-radius: 8px;">
                <img
                  src={state.qrDataUrl}
                  alt="QR Code for access"
                  style="max-width: 256px; width: 100%; height: auto;"
                />
              </div>
              <small className="form-help">
                Scan this QR code to quickly access the{" "}
                {props.targetLabel.toLowerCase()}
              </small>
            </div>
          )}

          <div className="stream-key-actions">
            <button
              className="btn btn-primary"
              onClick={() => copyCode(state)}
              disabled={state.copyCodeSuccess}
            >
              {state.copyCodeSuccess ? "✓ Copied Code!" : "Copy Code"}
            </button>
            <button
              className="btn btn-secondary"
              onClick={() => copyShareUrl(state)}
              disabled={state.copyUrlSuccess}
            >
              {state.copyUrlSuccess ? "✓ Copied URL!" : "Copy URL"}
            </button>
            {state.qrDataUrl && (
              <button
                className="btn btn-secondary"
                onClick={() => downloadQRCode(state)}
              >
                Download QR Code
              </button>
            )}
          </div>
        </div>
      ) : (
        // Form view - collect code parameters
        <div>
          <div className="form-group">
            <label>{props.targetLabel}</label>
            <div className="stream-key-room-name">{props.targetName}</div>
          </div>

          <div className="form-group">
            <label htmlFor="code-duration">Code Duration *</label>
            <select
              id="code-duration"
              className="form-input"
              {...vlens.attrsBindInput(vlens.ref(state, "duration"))}
              disabled={state.isSubmitting}
            >
              <option value="1h">1 hour</option>
              <option value="24h">24 hours</option>
              <option value="7d">7 days</option>
              <option value="30d">30 days</option>
              <option value="never">Never expires</option>
            </select>
            <small className="form-help">
              How long the access code will remain valid
            </small>
          </div>

          <div className="form-group">
            <label htmlFor="code-max-viewers">Max Viewers *</label>
            <input
              id="code-max-viewers"
              type="number"
              className="form-input"
              min="1"
              placeholder="e.g., 30"
              {...vlens.attrsBindInput(vlens.ref(state, "maxViewers"))}
              disabled={state.isSubmitting}
            />
            <small className="form-help">
              Maximum number of viewers who can use this code
            </small>
          </div>

          <div className="form-group">
            <label htmlFor="code-label">Label (Optional)</label>
            <input
              id="code-label"
              type="text"
              className="form-input"
              placeholder="e.g., Class Section A, Guest Viewers"
              {...vlens.attrsBindInput(vlens.ref(state, "label"))}
              disabled={state.isSubmitting}
            />
            <small className="form-help">
              Optional label to help you identify this code later
            </small>
          </div>
        </div>
      )}
    </Modal>
  );
}
