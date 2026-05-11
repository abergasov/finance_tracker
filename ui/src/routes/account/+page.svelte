<script lang="ts">
	import { onMount } from "svelte";
	import {
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
	let error = "";

	onMount(async () => {
		const storedSession = loadSession();
		if (!storedSession) {
			loading = false;
			return;
		}

		const result = await fetchCurrentUser(storedSession.token);
		if (!result.ok) {
			if (result.errorKind === "auth") {
				clearSession();
				error = "Your session expired. Sign in again.";
			} else {
				error = "Could not load your account. Please try again.";
			}
			loading = false;
			return;
		}

		session = {
			token: storedSession.token,
			user: result.data.user,
		};
		categories = result.data.categories;
		saveSession(session);
		loading = false;
	});

	async function refreshCategories(): Promise<void> {
		if (!session) return;

		try {
			const result = await fetchCurrentUser(session.token);
			if (result.ok) {
				categories = result.data.categories;
			}
		} catch {
			error = "Unable to refresh categories right now. Please try again.";
		}
	}
</script>

<svelte:head>
	<title>Account – Finance Tracker</title>
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
			<p class="eyebrow">Account</p>
			<h1>Expense categories</h1>
			<a class="back-link" href="/">← Back to home</a>
		</div>

		<div class="card categories-card">
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
		</div>
	</div>
{:else}
	<div class="shell">
		<div class="card">
			<p class="eyebrow">Account</p>
			<h1>Not signed in</h1>
			<p>Please sign in to manage your categories.</p>
			<a class="button" href="/">← Back to home</a>
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

	.eyebrow {
		margin: 0 0 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-size: 0.75rem;
		color: var(--text-muted);
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

	.back-link {
		display: inline-block;
		font-size: 0.9rem;
		color: var(--text-muted);
		text-decoration: none;
	}

	.back-link:hover {
		color: var(--text-primary);
	}

	.button {
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
