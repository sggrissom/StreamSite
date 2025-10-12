import * as preact from "preact";
import "./Modal-styles";

export type ModalProps = {
  isOpen: boolean;
  title: string;
  onClose: () => void;
  children: preact.ComponentChild;
  footer?: preact.ComponentChild;
  error?: string;
  size?: "default" | "large";
};

export function Modal(props: ModalProps): preact.ComponentChild {
  if (!props.isOpen) {
    return null;
  }

  const sizeClass = props.size === "large" ? "modal-content-large" : "";

  return (
    <div className="modal-overlay" onClick={props.onClose}>
      <div
        className={`modal-content ${sizeClass}`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="modal-header">
          <h2 className="modal-title">{props.title}</h2>
          <button className="modal-close" onClick={props.onClose}>
            ×
          </button>
        </div>

        <div className="modal-body">
          {props.error && <div className="error-message">{props.error}</div>}
          {props.children}
        </div>

        {props.footer && <div className="modal-footer">{props.footer}</div>}
      </div>
    </div>
  );
}
