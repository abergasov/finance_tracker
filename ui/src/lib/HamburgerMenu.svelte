<script lang="ts">
	import { onMount } from "svelte";
	import { clearSession, loadSession } from "$lib/auth";

	let open = false;
	let isSignedIn = false;
	let menuEl: HTMLDivElement | null = null;

	onMount(() => {
		isSignedIn = loadSession() !== null;
	});

	function toggleMenu() {
		// Re-check auth state each time the menu is opened so it stays consistent
		// even if a page-level signout was performed before the menu was used.
		isSignedIn = loadSession() !== null;
		open = !open;
		
		// Focus first menu item when opening
		if (open) {
			setTimeout(() => {
				const firstItem = menuEl?.querySelector<HTMLElement>('[role="menuitem"]');
				firstItem?.focus();
			}, 0);
		}
	}

	function closeMenu() {
		open = false;
	}

	function signOut() {
		clearSession();
		open = false;
		// Full navigation forces all in-memory state to reset across all pages.
		window.location.href = "/";
	}

	function handleWindowPointerDown(e: PointerEvent) {
		if (open && menuEl !== null && !menuEl.contains(e.target as Node)) {
			open = false;
		}
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (!open) return;

		switch (e.key) {
			case "Escape":
				e.preventDefault();
				open = false;
				// Return focus to the hamburger button
				menuEl?.querySelector<HTMLButtonElement>(".hamburger-btn")?.focus();
				break;
			case "ArrowDown":
			case "ArrowUp":
				e.preventDefault();
				// Get all currently visible menu items
				const menuItems = menuEl?.querySelectorAll<HTMLElement>('[role="menuitem"]');
				if (!menuItems || menuItems.length === 0) return;
				
				const currentIndex = Array.from(menuItems).indexOf(document.activeElement as HTMLElement);
				let nextIndex: number;
				
				if (e.key === "ArrowDown") {
					// Move to next item, or wrap to first
					nextIndex = currentIndex < 0 || currentIndex >= menuItems.length - 1 ? 0 : currentIndex + 1;
				} else {
					// Move to previous item, or wrap to last
					nextIndex = currentIndex <= 0 ? menuItems.length - 1 : currentIndex - 1;
				}
				
				menuItems[nextIndex]?.focus();
				break;
		}
	}
</script>

<svelte:window on:pointerdown={handleWindowPointerDown} on:keydown={handleKeyDown} />

<div class="hamburger-root" bind:this={menuEl}>
	<button
		class="hamburger-btn"
		on:click={toggleMenu}
		aria-label={open ? "Close navigation menu" : "Open navigation menu"}
		aria-expanded={open}
		aria-haspopup="true"
		aria-controls="nav-dropdown"
	>
		<span class="bar"></span>
		<span class="bar"></span>
		<span class="bar"></span>
	</button>

	{#if open}
		<div class="dropdown" id="nav-dropdown" role="menu">
			<a 
				class="menu-item" 
				href="/account" 
				on:click={closeMenu} 
				role="menuitem"
			>
				Account
			</a>
			{#if isSignedIn}
				<button 
					class="menu-item menu-item--btn" 
					on:click={signOut} 
					role="menuitem"
				>
					Sign out
				</button>
			{/if}
		</div>
	{/if}
</div>

<style>
	.hamburger-root {
		position: fixed;
		top: 0.75rem;
		/* Sits to the left of the ThemeSwitcher (2.5rem wide) with a 0.5rem gap */
		right: 3.75rem;
		z-index: 1000;
	}

	.hamburger-btn {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 0.3rem;
		width: 2.5rem;
		height: 2.5rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		background: var(--bg-card);
		cursor: pointer;
		box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
		transition:
			background 0.2s,
			border-color 0.2s;
	}

	.hamburger-btn:hover {
		background: var(--bg-hover);
	}

	.bar {
		display: block;
		width: 1.1rem;
		height: 2px;
		background: var(--text-primary);
		border-radius: 1px;
		transition: background 0.2s;
	}

	.dropdown {
		position: absolute;
		top: calc(100% + 0.4rem);
		right: 0;
		min-width: 10rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		background: var(--bg-card);
		box-shadow: 0 4px 16px var(--shadow-card);
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.menu-item {
		display: block;
		padding: 0.65rem 1rem;
		font: inherit;
		font-size: 0.9rem;
		color: var(--text-primary);
		background: transparent;
		text-decoration: none;
		border: none;
		text-align: left;
		cursor: pointer;
		transition: background 0.15s;
	}

	.menu-item:hover {
		background: var(--bg-hover);
	}

	.menu-item--btn {
		width: 100%;
		border-top: 1px solid var(--border);
	}
</style>
