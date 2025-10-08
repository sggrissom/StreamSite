import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "./home-styles";

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
      <main className="home-container">
        <h1 className="home-title">Deploy test</h1>
        <p className="home-description">we are so back</p>
        <div className="stream-link-container">
          <a href="/stream" className="stream-link">
            Watch Live Stream
          </a>
        </div>
      </main>
      <Footer />
    </div>
  );
}
