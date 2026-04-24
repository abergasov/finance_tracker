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

		const user = await fetchCurrentUser(storedSession.token);
		if (!user) {
			clearSession();
			error = "Your session expired. Sign in again.";
			loading = false;
			return;
		}

		session = {
			token: storedSession.token,
			user,
		};
		saveSession(session);
		loading = false;
	});

	function signOut() {
		clearSession();
		session = null;
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
	:global(body) {
		margin: 0;
		font-family: Arial, sans-serif;
		background: #f3f4f6;
		color: #111827;
	}

	.shell {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: 1.5rem;
	}

	.card {
		width: min(100%, 28rem);
		padding: 2rem;
		border-radius: 1rem;
		background: #ffffff;
		box-shadow: 0 20px 45px rgba(15, 23, 42, 0.08);
	}

	.eyebrow {
		margin: 0 0 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-size: 0.75rem;
		color: #4b5563;
	}

	h1 {
		margin: 0 0 0.75rem;
		font-size: 2rem;
	}

	p {
		margin: 0 0 1rem;
	}

	.button,
	button {
		display: inline-flex;
		justify-content: center;
		align-items: center;
		padding: 0.8rem 1rem;
		border: none;
		border-radius: 0.75rem;
		background: #111827;
		color: #ffffff;
		font: inherit;
		text-decoration: none;
		cursor: pointer;
	}

	.error {
		color: #b91c1c;
		margin-top: 1rem;
	}
</style>
