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

const storageKey = "finance-tracker-auth";

export function loadSession(): AuthSession | null {
	if (typeof window === "undefined") {
		return null;
	}

	const raw = window.localStorage.getItem(storageKey);
	if (!raw) {
		return null;
	}

	try {
		return JSON.parse(raw) as AuthSession;
	} catch {
		window.localStorage.removeItem(storageKey);
		return null;
	}
}

export function saveSession(session: AuthSession): void {
	if (typeof window === "undefined") {
		return;
	}

	window.localStorage.setItem(storageKey, JSON.stringify(session));
}

export function clearSession(): void {
	if (typeof window === "undefined") {
		return;
	}

	window.localStorage.removeItem(storageKey);
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
