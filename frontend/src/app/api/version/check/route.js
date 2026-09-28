import http from "http";

export async function POST() {
  try {
    const res = await new Promise((resolve, reject) => {
      const req = http.request(
        "http://127.0.0.1:20128/api/version/check",
        { method: "POST", timeout: 15000 },
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
  } catch (err) {
    return Response.json({ success: false, error: err.message }, { status: 500 });
  }
}

export async function GET() {
  return POST();
}
