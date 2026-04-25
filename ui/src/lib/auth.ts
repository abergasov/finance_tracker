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

export type ExpensesCategory = {
	id: number;
	name: string;
	children?: ExpensesCategory[];
};

export type UserExpenses = {
	mandatoryExpenses: ExpensesCategory;
	optionalExpenses: ExpensesCategory;
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

export type CurrentUserResult = {
	user: AuthUser;
	categories: UserExpenses | null;
};

export async function fetchCurrentUser(token: string): Promise<CurrentUserResult | null> {
	const response = await fetch(buildBackendURL("/api/auth/me"), {
		headers: {
			Authorization: `Bearer ${token}`,
		},
	});
	if (!response.ok) {
		return null;
	}

	const data = (await response.json()) as {
		user: AuthUser;
		user_expenses_categories: UserExpenses | null;
	};
	return {
		user: data.user,
		categories: data.user_expenses_categories ?? null,
	};
}

export async function createCategory(
	token: string,
	parentId: number,
	name: string,
): Promise<{ id: number } | null> {
	const response = await fetch(buildBackendURL("/api/auth/category"), {
		method: "POST",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({ parent_id: parentId, name }),
	});
	if (!response.ok) {
		return null;
	}
	return (await response.json()) as { id: number };
}

export async function updateCategory(
	token: string,
	id: number,
	name: string,
): Promise<boolean> {
	const response = await fetch(buildBackendURL(`/api/auth/category/${id}`), {
		method: "PUT",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({ name }),
	});
	return response.ok;
}

export async function deleteCategory(token: string, id: number): Promise<boolean> {
	const response = await fetch(buildBackendURL(`/api/auth/category/${id}`), {
		method: "DELETE",
		headers: {
			Authorization: `Bearer ${token}`,
		},
	});
	return response.ok;
}

function ensureTrailingSlash(value: string): string {
	return value.endsWith("/") ? value : `${value}/`;
}
