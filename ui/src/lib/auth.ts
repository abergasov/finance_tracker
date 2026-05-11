import { env } from "$env/dynamic/public";

export type AuthUser = {
	id: string;
	email: string;
	locale: string;
	name: string;
	default_currency: string;
};

export type AuthSession = {
	token: string;
	user: AuthUser;
};

export type ExpensesCategory = {
	id: number;
	name: string;
	color: string;
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
	supportedCurrencies: string[];
	categories: UserExpenses | null;
};

export type FetchCurrentUserError = "auth" | "network";

export type FetchCurrentUserOutcome =
	| { ok: true; data: CurrentUserResult }
	| { ok: false; errorKind: FetchCurrentUserError };

export async function fetchCurrentUser(token: string): Promise<FetchCurrentUserOutcome> {
	const response = await fetch(buildBackendURL("/api/v1/me"), {
		headers: {
			Authorization: `Bearer ${token}`,
		},
	});
	if (!response.ok) {
		const isAuthError = response.status === 401 || response.status === 403;
		return { ok: false, errorKind: isAuthError ? "auth" : "network" };
	}

	const data = (await response.json()) as {
		user: AuthUser;
		supported_currencies: string[];
		user_expenses_categories: UserExpenses | null;
	};
	return {
		ok: true,
		data: {
			user: data.user,
			supportedCurrencies: data.supported_currencies ?? [],
			categories: data.user_expenses_categories ?? null,
		},
	};
}

export async function createCategory(
	token: string,
	parentId: number,
	name: string,
	color?: string,
): Promise<{ id: number } | null> {
	const response = await fetch(buildBackendURL("/api/v1/category"), {
		method: "POST",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({ parent_id: parentId, name, color }),
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
	color?: string,
): Promise<boolean> {
	const response = await fetch(buildBackendURL(`/api/v1/category/${id}`), {
		method: "PUT",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({ name, color }),
	});
	return response.ok;
}

export async function deleteCategory(token: string, id: number): Promise<boolean> {
	const response = await fetch(buildBackendURL(`/api/v1/category/${id}`), {
		method: "DELETE",
		headers: {
			Authorization: `Bearer ${token}`,
		},
	});
	return response.ok;
}

export type UpdateCurrencyError = "auth" | "error";

export type UpdateCurrencyOutcome = { ok: true } | { ok: false; errorKind: UpdateCurrencyError };

export async function updateDefaultCurrency(token: string, currency: string): Promise<UpdateCurrencyOutcome> {
	try {
		const response = await fetch(buildBackendURL("/api/v1/me/currency"), {
			method: "PUT",
			headers: {
				Authorization: `Bearer ${token}`,
				"Content-Type": "application/json",
			},
			body: JSON.stringify({ currency }),
		});
		if (!response.ok) {
			const isAuthError = response.status === 401 || response.status === 403;
			return { ok: false, errorKind: isAuthError ? "auth" : "error" };
		}
		return { ok: true };
	} catch {
		return { ok: false, errorKind: "error" };
	}
}

function ensureTrailingSlash(value: string): string {
	return value.endsWith("/") ? value : `${value}/`;
}
