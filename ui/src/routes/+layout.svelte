<script lang="ts">
	import ThemeSwitcher from "$lib/ThemeSwitcher.svelte";
</script>

<ThemeSwitcher />
<slot />

<style>
	/*
	 * Global CSS custom properties for theming.
	 * Light values are the defaults; dark values override via [data-theme="dark"].
	 * The app.html inline script ensures data-theme is set before hydration
	 * to prevent any flash on first paint.
	 */
	:global(:root) {
		--bg-page: #f3f4f6;
		--bg-card: #ffffff;
		--bg-hover: #f9fafb;
		--text-primary: #111827;
		--text-muted: #4b5563;
		--btn-bg: #111827;
		--btn-text: #ffffff;
		--border: #e5e7eb;
		--shadow-card: rgba(15, 23, 42, 0.08);
		--error: #b91c1c;
	}

	:global(html[data-theme="dark"]) {
		--bg-page: #111827;
		--bg-card: #1f2937;
		--bg-hover: #374151;
		--text-primary: #f9fafb;
		--text-muted: #9ca3af;
		--btn-bg: #f9fafb;
		--btn-text: #111827;
		--border: #374151;
		--shadow-card: rgba(0, 0, 0, 0.35);
		--error: #fca5a5;
	}

	:global(body) {
		margin: 0;
		font-family: Arial, sans-serif;
		background: var(--bg-page);
		color: var(--text-primary);
		transition:
			background 0.2s,
			color 0.2s;
	}
</style>
