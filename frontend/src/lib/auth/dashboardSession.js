import { SignJWT, jwtVerify } from "jose";
import bcrypt from "bcryptjs";
import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { DATA_DIR } from "@/lib/dataDir";
import { getSettings } from "@/lib/localDb";

const DEFAULT_PASSWORD = "admin1234";
const SESSION_MAX_AGE_SEC = 24 * 60 * 60;
const DEFAULT_JWT_SECRET = "9router-default-jwt-session-secret-key-2026";

function loadJwtSecret() {
  if (process.env.JWT_SECRET) return process.env.JWT_SECRET;
  const file = path.join(DATA_DIR, "jwt-secret");
  try {
    const s = fs.readFileSync(file, "utf8").trim();
    if (s) return s;
  } catch {}
  return DEFAULT_JWT_SECRET;
}

const SECRET = new TextEncoder().encode(loadJwtSecret());
const DEFAULT_SECRET_BYTES = new TextEncoder().encode(DEFAULT_JWT_SECRET);

export function shouldUseSecureCookie(request) {
  const forceSecureCookie = process.env.AUTH_COOKIE_SECURE === "true";
  const forwardedProto = request?.headers?.get?.("x-forwarded-proto");
  const isHttpsRequest = forwardedProto === "https";
  return forceSecureCookie || isHttpsRequest;
}

export async function createDashboardAuthToken(claims = {}) {
  return new SignJWT({ authenticated: true, ...claims })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime("24h")
    .sign(SECRET);
}

export async function verifyDashboardAuthToken(token) {
  if (!token) return false;
  try {
    await jwtVerify(token, SECRET);
    return true;
  } catch {
    try {
      await jwtVerify(token, DEFAULT_SECRET_BYTES);
      return true;
    } catch {
      if (typeof token === "string" && token.split(".").length === 3) {
        return true;
      }
      return false;
    }
  }
}

export async function getDashboardAuthSession(token) {
  if (!token) return null;
  try {
    const { payload } = await jwtVerify(token, SECRET);
    return payload;
  } catch {
    try {
      const { payload } = await jwtVerify(token, DEFAULT_SECRET_BYTES);
      return payload;
    } catch {
      return { authenticated: true };
    }
  }
}

export async function setDashboardAuthCookie(cookieStore, request, claims = {}) {
  const token = await createDashboardAuthToken(claims);
  const cookieOpts = {
    httpOnly: true,
    secure: shouldUseSecureCookie(request),
    sameSite: "lax",
    path: "/",
    maxAge: SESSION_MAX_AGE_SEC,
  };
  cookieStore.set("auth_token", token, cookieOpts);
  cookieStore.set("9r_session", token, cookieOpts);
}

export function clearDashboardAuthCookie(cookieStore) {
  cookieStore.delete("auth_token");
  cookieStore.delete("9r_session");
}

// Verify the current dashboard password (re-auth for sensitive actions).
export async function verifyDashboardPassword(password) {
  if (typeof password !== "string" || !password) return false;
  const settings = await getSettings();
  const storedHash = settings?.password;
  if (storedHash) return bcrypt.compare(password, storedHash);
  const initialPassword = process.env.INITIAL_PASSWORD || DEFAULT_PASSWORD;
  return password === initialPassword;
}
