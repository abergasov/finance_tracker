<script lang="ts">
	import { onMount } from "svelte";
	import { theme } from "$lib/theme";

	// Guard against rendering the icon before init() has synced the store.
	// Without this, the store default ("light") would briefly show 🌙 on a
	// dark-theme page before onMount fires.
	let mounted = false;

	onMount(() => {
		theme.init();
		mounted = true;
	});
</script>

<!-- Fixed position keeps it accessible on every route at every breakpoint. -->
<button class="theme-toggle" on:click={theme.toggle} disabled={!mounted} aria-label="Toggle theme">
	{#if mounted}
		{#if $theme === "dark"}
			☀️
		{:else}
			🌙
		{/if}
	{/if}
</button>

<style>
	.theme-toggle {
		position: fixed;
		top: 0.75rem;
		right: 0.75rem;
		z-index: 1000;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 2.5rem;
		height: 2.5rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		background: var(--bg-card);
		color: var(--text-primary);
		font-size: 1.1rem;
		cursor: pointer;
		box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
		transition:
			background 0.2s,
			border-color 0.2s;
	}

	.theme-toggle:hover {
		background: var(--bg-hover);
	}

	.theme-toggle:disabled {
		cursor: default;
		pointer-events: none;
	}
</style>
