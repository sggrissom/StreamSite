import * as preact from "preact";
import * as rpc from "vlens/rpc";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./dashboard-styles";

type Data = {};

export async function fetch(route: string, prefix: string) {
  return rpc.ok<Data>({});
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  return (
    <div>
      <Header />
      <main className="dashboard-container">
        <div className="dashboard-content">
          <h1 className="dashboard-title">Welcome to Stream</h1>
          <p className="dashboard-description">
            You're all set! Ready to watch the stream?
          </p>
          <div className="dashboard-actions">
            <a href="/stream" className="btn btn-primary btn-large">
              Watch Live Stream
            </a>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
