<script lang="ts">
	import { onMount } from "svelte";
	import { goto } from "$app/navigation";
	import { saveSession } from "$lib/auth";

	let message = "Completing sign-in…";

	onMount(async () => {
		const hashParams = new URLSearchParams(window.location.hash.replace(/^#/, ""));

		// Remove the fragment from browser history immediately so the token is
		// not visible in the URL bar or retained in history.
		history.replaceState(null, "", window.location.pathname + window.location.search);

		const error = hashParams.get("error");
		if (error) {
			message = error;
			return;
		}

		const token = hashParams.get("token");
		const id = hashParams.get("id");
		const email = hashParams.get("email");
		if (!token || !id || !email) {
			message = "Sign-in response is missing required fields.";
			return;
		}

		saveSession({
			token,
			user: {
				id,
				email,
				name: hashParams.get("name") ?? "",
				locale: hashParams.get("locale") ?? "",
			},
		});
		await goto("/");
	});
</script>

<div class="shell">
	<div class="card">
		<p>{message}</p>
	</div>
</div>

<style>
	.shell {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: 1.5rem;
	}

	.card {
		padding: 2rem;
		border-radius: 1rem;
		background: var(--bg-card);
		box-shadow: 0 20px 45px var(--shadow-card);
	}
</style>
