<script lang="ts">
	import { onMount } from "svelte";
	import {
		buildGoogleLoginURL,
		clearSession,
		fetchCurrentUser,
		loadSession,
		saveSession,
		type AuthSession,
	} from "$lib/auth";

	let loading = true;
	let session: AuthSession | null = null;
	let error = "";
	const googleLoginURL = buildGoogleLoginURL();

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
		saveSession(session);
		loading = false;
	});
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
