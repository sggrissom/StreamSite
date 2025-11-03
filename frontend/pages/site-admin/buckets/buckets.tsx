import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Header, Footer } from "../../../layout";
import "../../../styles/global";
import "./buckets-styles";

type Data = server.ListBucketsResponse;

export async function fetch(route: string, prefix: string) {
  return server.ListBuckets({});
}

type PageState = {
  selectedBucket: string;
  bucketData: server.GetBucketDataResponse | null;
  isLoading: boolean;
  error: string;
};

const usePageState = vlens.declareHook(
  (): PageState => ({
    selectedBucket: "",
    bucketData: null,
    isLoading: false,
    error: "",
  }),
);

async function loadBucketData(state: PageState, bucketName: string) {
  state.selectedBucket = bucketName;
  state.isLoading = true;
  state.error = "";
  state.bucketData = null;
  vlens.scheduleRedraw();

  const [resp, err] = await server.GetBucketData({ bucketName });

  state.isLoading = false;

  if (err) {
    state.error = err || "Failed to load bucket data";
    vlens.scheduleRedraw();
    return;
  }

  state.bucketData = resp;
  vlens.scheduleRedraw();
}

function formatValue(value: any): string {
  if (value === null || value === undefined) {
    return "null";
  }
  if (typeof value === "object") {
    return JSON.stringify(value, null, 2);
  }
  return String(value);
}

export function view(route: string, prefix: string, data: Data) {
  const state = usePageState();

  // Check if user is authenticated as site admin
  if (!data || !data.buckets) {
    return (
      <div>
        <Header />
        <main className="bucket-viewer-container">
          <div className="bucket-viewer-error">
            <h1>Access Denied</h1>
            <p>Only site administrators can access the bucket viewer.</p>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  return (
    <div>
      <Header />
      <main className="bucket-viewer-container">
        <div className="bucket-viewer-header">
          <div style={{ marginBottom: "1rem" }}>
            <a href="/site-admin" className="btn btn-secondary btn-sm">
              ← Back to Site Admin
            </a>
          </div>
          <h1>Database Bucket Viewer</h1>
          <p className="bucket-viewer-subtitle">
            Inspect database buckets and indexes
          </p>
        </div>

        <div className="bucket-viewer-content">
          {/* Left Sidebar: Bucket List */}
          <div className="bucket-list-sidebar">
            <h2>Buckets ({data.buckets.filter((b) => !b.isIndex).length})</h2>
            <div className="bucket-list">
              {data.buckets
                .filter((bucket) => !bucket.isIndex)
                .map((bucket) => (
                  <div
                    key={bucket.name}
                    className={
                      "bucket-list-item" +
                      (state.selectedBucket === bucket.name ? " active" : "")
                    }
                    onClick={() => loadBucketData(state, bucket.name)}
                  >
                    <div className="bucket-name">{bucket.name}</div>
                    <div className="bucket-description">
                      {bucket.description}
                    </div>
                    <div className="bucket-meta">
                      {bucket.keyType} → {bucket.valueType}
                    </div>
                  </div>
                ))}
            </div>

            <h2>Indexes ({data.buckets.filter((b) => b.isIndex).length})</h2>
            <div className="bucket-list">
              {data.buckets
                .filter((bucket) => bucket.isIndex)
                .map((bucket) => (
                  <div
                    key={bucket.name}
                    className={
                      "bucket-list-item" +
                      (state.selectedBucket === bucket.name ? " active" : "")
                    }
                    onClick={() => loadBucketData(state, bucket.name)}
                  >
                    <div className="bucket-name">{bucket.name}</div>
                    <div className="bucket-description">
                      {bucket.description}
                    </div>
                    <div className="bucket-meta">
                      {bucket.keyType} → {bucket.valueType}
                    </div>
                  </div>
                ))}
            </div>
          </div>

          {/* Main Area: Bucket Data */}
          <div className="bucket-data-area">
            {!state.selectedBucket && (
              <div className="bucket-data-empty">
                <p>Select a bucket or index to view its contents</p>
              </div>
            )}

            {state.isLoading && (
              <div className="bucket-data-loading">
                <p>Loading...</p>
              </div>
            )}

            {state.error && (
              <div className="bucket-data-error">
                <p>Error: {state.error}</p>
              </div>
            )}

            {state.bucketData && !state.isLoading && (
              <div className="bucket-data-container">
                <div className="bucket-data-header">
                  <h2>{state.selectedBucket}</h2>
                  <p className="bucket-data-count">
                    {state.bucketData.total} entries
                  </p>
                </div>

                {state.bucketData.total === 0 ? (
                  <div className="bucket-data-empty">
                    <p>No entries in this bucket</p>
                  </div>
                ) : (
                  <div className="bucket-data-table-container">
                    <table className="bucket-data-table">
                      <thead>
                        <tr>
                          <th>Key</th>
                          <th>Value</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.bucketData.entries.map((entry, index) => (
                          <tr key={index}>
                            <td className="bucket-data-key">
                              {formatValue(entry.key)}
                            </td>
                            <td className="bucket-data-value">
                              <pre>{formatValue(entry.value)}</pre>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
