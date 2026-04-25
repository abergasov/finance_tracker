<script lang="ts">
	import { onMount } from "svelte";
	import {
		buildGoogleLoginURL,
		clearSession,
		fetchCurrentUser,
		loadSession,
		saveSession,
		type AuthSession,
		type UserExpenses,
	} from "$lib/auth";
	import CategoryNode from "$lib/CategoryNode.svelte";

	let loading = true;
	let session: AuthSession | null = null;
	let categories: UserExpenses | null = null;
	let categoriesExpanded = true;
	let error = "";
	const googleLoginURL = buildGoogleLoginURL();

	onMount(async () => {
		const storedSession = loadSession();
		if (!storedSession) {
			loading = false;
			return;
		}

		const result = await fetchCurrentUser(storedSession.token);
		if (!result) {
			clearSession();
			error = "Your session expired. Sign in again.";
			loading = false;
			return;
		}

		session = {
			token: storedSession.token,
			user: result.user,
		};
		categories = result.categories;
		saveSession(session);
		loading = false;
	});

	async function refreshCategories() {
		if (!session) return;
		const result = await fetchCurrentUser(session.token);
		if (result) {
			categories = result.categories;
		}
	}

	function signOut() {
		clearSession();
		session = null;
		categories = null;
		error = "";
	}
</script>

<svelte:head>
	<title>Finance Tracker</title>
</svelte:head>

{#if loading}
	<div class="shell">
		<div class="card">
			<p>Loading…</p>
		</div>
	</div>
{:else if session}
	<div class="shell">
		<div class="card">
			<p class="eyebrow">Authenticated home</p>
			<h1>Welcome, {session.user.name || session.user.email}</h1>
			<p>{session.user.email}</p>
			{#if session.user.locale}
				<p>Locale: {session.user.locale}</p>
			{/if}
			<button on:click={signOut}>Sign out</button>
		</div>

		<!-- Categories card -->
		<div class="card categories-card">
			<button
				class="card-header-btn"
				on:click={() => (categoriesExpanded = !categoriesExpanded)}
				aria-expanded={categoriesExpanded}
			>
				<span class="eyebrow">Expense categories</span>
				<span class="chevron">{categoriesExpanded ? "▾" : "▸"}</span>
			</button>

			{#if categoriesExpanded}
				{#if !categories || (!categories.mandatoryExpenses?.id && !categories.optionalExpenses?.id)}
					<p class="muted">No categories yet.</p>
				{:else}
					<div class="tree">
						{#if categories.mandatoryExpenses?.id}
							<CategoryNode
								node={categories.mandatoryExpenses}
								isRoot={true}
								token={session.token}
								onRefresh={refreshCategories}
							/>
						{/if}
						{#if categories.optionalExpenses?.id}
							<CategoryNode
								node={categories.optionalExpenses}
								isRoot={true}
								token={session.token}
								onRefresh={refreshCategories}
							/>
						{/if}
					</div>
				{/if}
			{/if}
		</div>
	</div>
{:else}
	<div class="shell">
		<div class="card">
			<p class="eyebrow">Finance Tracker</p>
			<h1>Sign in</h1>
			<p>Continue with Google to reach the app home page.</p>
			<a class="button" href={googleLoginURL}>Continue with Google</a>
			{#if error}
				<p class="error">{error}</p>
			{/if}
		</div>
	</div>
{/if}

<style>
	.shell {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 1.5rem;
		gap: 1rem;
	}

	.card {
		width: min(100%, 28rem);
		padding: 2rem;
		border-radius: 1rem;
		background: var(--bg-card);
		box-shadow: 0 20px 45px var(--shadow-card);
	}

	.categories-card {
		padding: 1rem 1.5rem;
	}

	.card-header-btn {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		background: none;
		border: none;
		padding: 0.5rem 0;
		cursor: pointer;
		color: inherit;
		font: inherit;
		border-radius: 0.5rem;
	}

	.card-header-btn:hover {
		background: var(--bg-hover, rgba(128, 128, 128, 0.08));
	}

	.chevron {
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.eyebrow {
		margin: 0 0 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	/* When eyebrow is inside the header button, remove bottom margin */
	.card-header-btn .eyebrow {
		margin: 0;
	}

	h1 {
		margin: 0 0 0.75rem;
		font-size: 2rem;
	}

	p {
		margin: 0 0 1rem;
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.85rem;
		margin: 0.5rem 0 0;
	}

	.tree {
		padding: 0.5rem 0 0;
	}

	.button,
	button {
		display: inline-flex;
		justify-content: center;
		align-items: center;
		padding: 0.8rem 1rem;
		border: none;
		border-radius: 0.75rem;
		background: var(--btn-bg);
		color: var(--btn-text);
		font: inherit;
		text-decoration: none;
		cursor: pointer;
	}

	.error {
		color: var(--error);
		margin-top: 1rem;
	}
</style>
