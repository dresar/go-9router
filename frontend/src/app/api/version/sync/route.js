import http from "http";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

export async function POST() {
  // Attempt proxy to Go gateway first
  try {
    const res = await new Promise((resolve, reject) => {
      const req = http.request(
        "http://127.0.0.1:20128/api/version/sync",
        { method: "POST", timeout: 25000 },
        (res) => {
          let data = "";
          res.on("data", (c) => (data += c));
          res.on("end", () => {
            try {
              resolve({ status: res.statusCode, body: JSON.parse(data) });
            } catch {
              resolve({ status: res.statusCode, body: { error: data } });
            }
          });
        }
      );
      req.on("error", reject);
      req.on("timeout", () => { req.destroy(); reject(new Error("timeout")); });
      req.end();
    });
    return Response.json(res.body, { status: res.status || 200 });
  } catch {
    // Fallback: direct git pull execution
    try {
      const { stdout, stderr } = await execAsync("git pull origin master");
      return Response.json({
        success: true,
        output: (stdout || stderr || "").trim(),
        message: "Git pull executed directly"
      });
    } catch (err) {
      return Response.json(
        { success: false, error: err.message },
        { status: 500 }
      );
    }
  }
}
