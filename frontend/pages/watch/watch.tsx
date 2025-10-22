import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";

type Data = {};

export async function fetch(route: string, prefix: string) {
  return rpc.ok<Data>({});
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  // Extract code from URL if present (e.g., /watch/12345)
  const urlCode = extractCodeFromRoute(route, prefix);

  // Redirect to home page, preserving the code in the URL if present
  const redirectTo = urlCode ? `/${urlCode}` : "/";
  core.setRoute(redirectTo);

  return <div></div>;
}

function extractCodeFromRoute(route: string, prefix: string): string {
  // Extract code from URL like /watch/12345
  const afterPrefix = route.substring(prefix.length);
  const parts = afterPrefix.split("/").filter((p) => p.length > 0);

  if (parts.length > 0) {
    const code = parts[0];
    // Validate it's a 5-digit code
    if (/^\d{5}$/.test(code)) {
      return code;
    }
  }

  return "";
}
