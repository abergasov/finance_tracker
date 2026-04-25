<script lang="ts">
	import type { ExpensesCategory } from '$lib/auth';
	import { createCategory, updateCategory, deleteCategory } from '$lib/auth';
	import CategoryNode from './CategoryNode.svelte';

	export let node: ExpensesCategory;
	export let isRoot: boolean = false;
	export let token: string;
	export let onRefresh: () => void;

	let expanded = true;
	let editMode = false;
	let editName = '';
	let addMode = false;
	let newName = '';
	let busy = false;
	let nodeError = '';

	function startEdit() {
		editName = node.name;
		editMode = true;
		nodeError = '';
	}

	function cancelEdit() {
		editMode = false;
		nodeError = '';
	}

	async function submitEdit() {
		if (!editName.trim()) return;
		busy = true;
		nodeError = '';
		const ok = await updateCategory(token, node.id, editName.trim());
		busy = false;
		if (!ok) {
			nodeError = 'Failed to rename category.';
			return;
		}
		editMode = false;
		onRefresh();
	}

	function startAdd() {
		newName = '';
		addMode = true;
		nodeError = '';
	}

	function cancelAdd() {
		addMode = false;
		nodeError = '';
	}

	async function submitAdd() {
		if (!newName.trim()) return;
		busy = true;
		nodeError = '';
		const result = await createCategory(token, node.id, newName.trim());
		busy = false;
		if (!result) {
			nodeError = 'Failed to create category.';
			return;
		}
		addMode = false;
		onRefresh();
	}

	async function handleDelete() {
		if (!confirm(`Delete "${node.name}" and all its subcategories?`)) return;
		busy = true;
		nodeError = '';
		const ok = await deleteCategory(token, node.id);
		busy = false;
		if (!ok) {
			nodeError = 'Failed to delete category.';
			return;
		}
		onRefresh();
	}

	$: hasChildren = node.children && node.children.length > 0;
</script>

<div class="node" class:root={isRoot}>
	<div class="node-row">
		<!-- Expand/collapse toggle shown only when there are children -->
		{#if hasChildren}
			<button
				class="toggle"
				aria-label={expanded ? 'Collapse' : 'Expand'}
				on:click={() => (expanded = !expanded)}
			>
				{expanded ? '▾' : '▸'}
			</button>
		{:else}
			<span class="toggle-placeholder"></span>
		{/if}

		{#if editMode}
			<input
				class="inline-input"
				bind:value={editName}
				on:keydown={(e) => {
					if (e.key === 'Enter') submitEdit();
					if (e.key === 'Escape') cancelEdit();
				}}
				disabled={busy}
			/>
			<button class="action-btn save" on:click={submitEdit} disabled={busy}>✓</button>
			<button class="action-btn cancel" on:click={cancelEdit} disabled={busy}>✕</button>
		{:else}
			<span class="node-name" class:root-name={isRoot}>{node.name}</span>
			<div class="actions">
				<button class="action-btn add" title="Add subcategory" on:click={startAdd} disabled={busy}>
					+
				</button>
				{#if !isRoot}
					<button class="action-btn edit" title="Rename" on:click={startEdit} disabled={busy}>
						✎
					</button>
					<button
						class="action-btn delete"
						title="Delete"
						on:click={handleDelete}
						disabled={busy}
					>
						✕
					</button>
				{/if}
			</div>
		{/if}
	</div>

	{#if nodeError}
		<p class="node-error">{nodeError}</p>
	{/if}

	{#if addMode}
		<div class="add-form">
			<input
				class="inline-input"
				placeholder="Category name"
				bind:value={newName}
				on:keydown={(e) => {
					if (e.key === 'Enter') submitAdd();
					if (e.key === 'Escape') cancelAdd();
				}}
				disabled={busy}
			/>
			<button class="action-btn save" on:click={submitAdd} disabled={busy || !newName.trim()}>
				Add
			</button>
			<button class="action-btn cancel" on:click={cancelAdd} disabled={busy}>✕</button>
		</div>
	{/if}

	{#if expanded && hasChildren}
		<div class="children">
			{#each node.children ?? [] as child (child.id)}
				<svelte:self node={child} isRoot={false} {token} {onRefresh} />
			{/each}
		</div>
	{/if}
</div>

<style>
	.node {
		padding: 0.1rem 0;
	}

	.node-row {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.3rem 0.4rem;
		border-radius: 0.4rem;
		min-height: 2rem;
	}

	.node-row:hover {
		background: var(--bg-hover, rgba(128, 128, 128, 0.08));
	}

	.toggle {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		font-size: 0.75rem;
		color: var(--text-muted);
		width: 1rem;
		text-align: center;
		flex-shrink: 0;
	}

	.toggle-placeholder {
		display: inline-block;
		width: 1rem;
		flex-shrink: 0;
	}

	.node-name {
		flex: 1;
		font-size: 0.9rem;
	}

	.root-name {
		font-weight: 600;
		font-size: 0.95rem;
	}

	.actions {
		display: flex;
		gap: 0.2rem;
		opacity: 0;
		transition: opacity 0.1s;
	}

	.node-row:hover .actions {
		opacity: 1;
	}

	.action-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.15rem 0.35rem;
		border-radius: 0.25rem;
		font-size: 0.8rem;
		color: var(--text-muted);
		line-height: 1;
	}

	.action-btn:hover:not(:disabled) {
		background: var(--btn-bg);
		color: var(--btn-text);
	}

	.action-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.action-btn.delete:hover:not(:disabled) {
		background: var(--error, #c0392b);
		color: #fff;
	}

	.action-btn.save:hover:not(:disabled) {
		background: var(--success, #27ae60);
		color: #fff;
	}

	.inline-input {
		flex: 1;
		font: inherit;
		font-size: 0.9rem;
		padding: 0.2rem 0.4rem;
		border: 1px solid var(--border, rgba(128, 128, 128, 0.4));
		border-radius: 0.3rem;
		background: var(--bg-card);
		color: inherit;
	}

	.add-form {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.3rem 0.4rem 0.3rem 1.4rem;
	}

	.children {
		padding-left: 1.2rem;
		border-left: 1px solid var(--border, rgba(128, 128, 128, 0.2));
		margin-left: 0.6rem;
	}

	.node-error {
		font-size: 0.78rem;
		color: var(--error, #c0392b);
		margin: 0.1rem 0.4rem 0.1rem 1.4rem;
	}
</style>
