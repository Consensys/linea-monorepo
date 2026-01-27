import { homedir } from "os";
import { join } from "path";

// File storage
export const DEFAULT_UPLOADS_DIR = join(homedir(), ".linea-verifier-ui");
export const UPLOADS_DIR = process.env.VERIFIER_UPLOADS_DIR || DEFAULT_UPLOADS_DIR;
export const SESSIONS_DIR = join(UPLOADS_DIR, "sessions");

// Session configuration
export const SESSION_EXPIRY_HOURS = parseInt(process.env.SESSION_EXPIRY_HOURS || "24", 10);
export const SESSION_EXPIRY_MS = SESSION_EXPIRY_HOURS * 60 * 60 * 1000;

// File limits
export const MAX_FILE_SIZE = parseInt(process.env.MAX_FILE_SIZE || "10485760", 10); // 10MB
export const MAX_SESSION_SIZE = parseInt(process.env.MAX_SESSION_SIZE || "52428800", 10); // 50MB

// Allowed file types
export const ALLOWED_CONFIG_EXTENSIONS = [".json", ".md"];
export const ALLOWED_SCHEMA_EXTENSIONS = [".json"];
export const ALLOWED_ARTIFACT_EXTENSIONS = [".json"];

// API paths
export const API_SESSION = "/api/session";
export const API_UPLOAD = "/api/upload";
export const API_VERIFY = "/api/verify";
