import { writable } from "svelte/store";
import { browser } from "$app/environment";

export type Theme = "light" | "dark";

const STORAGE_KEY = "theme";

function resolveInitialTheme(): Theme {
	if (!browser) return "light";
	// Prefer what the app.html script already applied to the DOM.
	const applied = document.documentElement.getAttribute("data-theme");
	if (applied === "dark" || applied === "light") return applied;
	// Fallback: localStorage, then OS preference.
	try {
		const saved = localStorage.getItem(STORAGE_KEY);
		if (saved === "dark" || saved === "light") return saved;
	} catch {
		// Ignore storage access failures and fall back to OS preference.
	}
	return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function applyTheme(t: Theme) {
	document.documentElement.setAttribute("data-theme", t);
	try {
		localStorage.setItem(STORAGE_KEY, t);
	} catch {
		// Ignore storage access failures; theme is still applied to the DOM.
	}
}

function createThemeStore() {
	const { subscribe, set, update } = writable<Theme>("light");

	return {
		subscribe,
		/** Call once on mount to sync the store with the DOM-applied theme.
		 *  Only writes localStorage when an explicit saved preference already
		 *  exists; an OS-derived fallback is NOT persisted here. */
		init() {
			if (!browser) return;
			const initial = resolveInitialTheme();
			// app.html already set data-theme; keep DOM in sync defensively
			// but do not call applyTheme(), which would persist an OS fallback.
			document.documentElement.setAttribute("data-theme", initial);
			set(initial);
		},
		toggle() {
			update((current) => {
				const next: Theme = current === "dark" ? "light" : "dark";
				applyTheme(next);
				return next;
			});
		},
	};
}

export const theme = createThemeStore();
