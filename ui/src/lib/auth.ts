import { env } from "$env/dynamic/public";

export type AuthUser = {
	id: string;
	email: string;
	locale: string;
	name: string;
};

export type AuthSession = {
	token: string;
	user: AuthUser;
};

// Session is kept in memory only (never written to localStorage/sessionStorage)
// so it cannot be exfiltrated by an XSS attack.  The trade-off is that the
// session is lost on a full page reload and the user must re-authenticate.
let _session: AuthSession | null = null;

export function loadSession(): AuthSession | null {
	return _session;
}

export function saveSession(session: AuthSession): void {
	_session = session;
}

export function clearSession(): void {
	_session = null;
}

export function buildBackendURL(path: string): string {
	const baseURL = env.PUBLIC_API_BASE_URL?.trim() ?? "";
	if (baseURL === "") {
		return path;
	}

	try {
		return new URL(path, ensureTrailingSlash(baseURL)).toString();
	} catch {
		return path;
	}
}

export function buildGoogleLoginURL(): string {
	return buildBackendURL("/api/auth/google/login");
}

export async function fetchCurrentUser(token: string): Promise<AuthUser | null> {
	const response = await fetch(buildBackendURL("/api/auth/me"), {
		headers: {
			Authorization: `Bearer ${token}`,
		},
	});
	if (!response.ok) {
		return null;
	}

	const data = (await response.json()) as { user: AuthUser };
	return data.user;
}

function ensureTrailingSlash(value: string): string {
	return value.endsWith("/") ? value : `${value}/`;
}
