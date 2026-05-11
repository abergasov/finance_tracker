<script lang="ts">
	import { createExpense, type ExpensesCategory, type UserExpenses } from '$lib/auth';

	export let open: boolean = false;
	export let token: string;
	export let defaultCurrency: string;
	export let supportedCurrencies: string[];
	export let categories: UserExpenses | null;
	export let onClose: () => void;
	export let onAuthFailure: () => void;

	// Flatten the category tree into a list of selectable (non-root) nodes.
	// Root nodes are mandatory and optional; every other node is selectable.
	function flattenNonRoot(node: ExpensesCategory, ancestors: string[]): { id: number; label: string }[] {
		const result: { id: number; label: string }[] = [];
		for (const child of node.children ?? []) {
			const path = [...ancestors, child.name];
			result.push({ id: child.id, label: path.join(' › ') });
			result.push(...flattenNonRoot(child, path));
		}
		return result;
	}

	$: flatCategories = (() => {
		const items: { id: number; label: string }[] = [];
		if (categories?.mandatoryExpenses?.id) {
			items.push(...flattenNonRoot(categories.mandatoryExpenses, ['Mandatory']));
		}
		if (categories?.optionalExpenses?.id) {
			items.push(...flattenNonRoot(categories.optionalExpenses, ['Optional']));
		}
		return items;
	})();

	let selectedCategoryId: number = 0;
	let selectedCurrency: string = defaultCurrency || 'USD';
	let amount: string = '';
	let submitting = false;
	let error = '';

	$: if (open) {
		// Reset form each time the modal opens.
		selectedCategoryId = flatCategories.length > 0 ? flatCategories[0].id : 0;
		selectedCurrency = defaultCurrency || 'USD';
		amount = '';
		error = '';
	}

	async function handleSubmit() {
		if (submitting) return;
		error = '';

		if (!selectedCategoryId) {
			error = 'Please select a category.';
			return;
		}
		if (!amount.trim()) {
			error = 'Please enter an amount.';
			return;
		}

		submitting = true;
		const result = await createExpense(token, selectedCategoryId, selectedCurrency, amount.trim());
		submitting = false;

		if (!result.ok) {
			if (result.errorKind === 'auth') {
				onAuthFailure();
				return;
			}
			error = result.error;
			return;
		}

		onClose();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) onClose();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}
</script>

{#if open}
	<div
		class="backdrop"
		role="dialog"
		tabindex="-1"
		aria-modal="true"
		aria-label="Add expense"
		on:click={handleBackdropClick}
		on:keydown={handleKeydown}
	>
		<div class="modal">
			<h2 class="modal-title">Add expense</h2>

			<label class="field-label" for="expense-category">Category</label>
			{#if flatCategories.length === 0}
				<p class="hint">No categories yet. Add some in the Account page first.</p>
			{:else}
				<select
					id="expense-category"
					class="field-select"
					bind:value={selectedCategoryId}
					disabled={submitting}
				>
					{#each flatCategories as cat (cat.id)}
						<option value={cat.id}>{cat.label}</option>
					{/each}
				</select>
			{/if}

			<label class="field-label" for="expense-currency">Currency</label>
			<select
				id="expense-currency"
				class="field-select"
				bind:value={selectedCurrency}
				disabled={submitting}
			>
				{#each supportedCurrencies as code}
					<option value={code}>{code}</option>
				{/each}
			</select>

			<label class="field-label" for="expense-amount">Amount</label>
			<input
				id="expense-amount"
				class="field-input"
				type="text"
				inputmode="decimal"
				placeholder="e.g. 12.34"
				bind:value={amount}
				disabled={submitting}
				on:keydown={(e) => { if (e.key === 'Enter') handleSubmit(); }}
			/>

			{#if error}
				<p class="error">{error}</p>
			{/if}

			<div class="actions">
				<button class="button secondary" on:click={onClose} disabled={submitting}>Cancel</button>
				<button
					class="button"
					on:click={handleSubmit}
					disabled={submitting || flatCategories.length === 0}
				>
					{submitting ? 'Saving…' : 'Save'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.45);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		z-index: 100;
	}

	.modal {
		background: var(--bg-card);
		border-radius: 1rem;
		padding: 2rem;
		width: min(100%, 24rem);
		box-shadow: 0 20px 45px var(--shadow-card);
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.modal-title {
		margin: 0 0 0.75rem;
		font-size: 1.3rem;
	}

	.field-label {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-muted);
		margin-top: 0.25rem;
	}

	.field-select,
	.field-input {
		padding: 0.6rem 0.75rem;
		border-radius: 0.5rem;
		border: 1px solid var(--border, #ccc);
		background: var(--bg-input, var(--bg-card));
		color: inherit;
		font: inherit;
		width: 100%;
		box-sizing: border-box;
	}

	.field-select:disabled,
	.field-input:disabled {
		opacity: 0.6;
	}

	.hint {
		font-size: 0.85rem;
		color: var(--text-muted);
		margin: 0;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
		margin-top: 0.75rem;
	}

	.button {
		display: inline-flex;
		justify-content: center;
		align-items: center;
		padding: 0.65rem 1.2rem;
		border: none;
		border-radius: 0.75rem;
		background: var(--btn-bg);
		color: var(--btn-text);
		font: inherit;
		cursor: pointer;
	}

	.button.secondary {
		background: transparent;
		color: var(--text-muted);
		border: 1px solid var(--border, #ccc);
	}

	.button:disabled {
		opacity: 0.6;
		cursor: default;
	}

	.error {
		color: var(--error);
		font-size: 0.88rem;
		margin: 0.25rem 0 0;
	}
</style>
