<script lang="ts">
	import { tick } from 'svelte';
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

	function pickCurrency(currencies: string[], preferred: string): string {
		if (preferred && currencies.includes(preferred)) {
			return preferred;
		}
		return currencies[0] ?? '';
	}

	let selectedCategoryId: number = 0;
	let selectedCurrency: string = defaultCurrency || 'USD';
	let currencySearch: string = '';
	let amount: string = '';
	let submitting = false;

	// Filter logic:
	//   0 chars  → full list
	//   1 char   → currencies containing that character (case-insensitive)
	//   2+ chars → normal case-insensitive substring match
	$: filteredCurrencies = (() => {
		if (currencySearch.length === 0) return supportedCurrencies;
		const q = currencySearch.toLowerCase();
		return supportedCurrencies.filter((c) => c.toLowerCase().includes(q));
	})();

	$: defaultSelectedCurrency = pickCurrency(supportedCurrencies, defaultCurrency || 'USD');

	// Keep selectedCurrency in sync when the filter hides the current selection.
	$: if (currencySearch.trim().length === 0) {
		selectedCurrency = pickCurrency(supportedCurrencies, selectedCurrency || defaultSelectedCurrency);
	} else if (filteredCurrencies.length > 0 && !filteredCurrencies.includes(selectedCurrency)) {
		selectedCurrency = filteredCurrencies[0];
	}

	$: submitCurrency =
		currencySearch.trim().length === 0
			? pickCurrency(supportedCurrencies, selectedCurrency || defaultSelectedCurrency)
			: pickCurrency(filteredCurrencies, selectedCurrency || defaultSelectedCurrency);

	let error = '';
	let dialogEl: HTMLDivElement | null = null;
	let previouslyFocusedEl: HTMLElement | null = null;
	let wasOpen = false;
	let pointerDownOutsideDialog = false;

	function getDeepActiveElement(): HTMLElement | null {
		let activeElement = document.activeElement as HTMLElement | null;
		while (
			activeElement &&
			'shadowRoot' in activeElement &&
			activeElement.shadowRoot?.activeElement instanceof HTMLElement
		) {
			activeElement = activeElement.shadowRoot.activeElement;
		}
		return activeElement instanceof HTMLElement ? activeElement : null;
	}

	async function focusDialog() {
		await tick();
		const firstFocusable = dialogEl?.querySelector<HTMLElement>(
			'button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
		);
		(firstFocusable ?? dialogEl)?.focus();
	}

	$: if (open && !wasOpen) {
		wasOpen = true;
		previouslyFocusedEl = getDeepActiveElement();

		// Reset form each time the modal opens.
		selectedCategoryId = flatCategories.length > 0 ? flatCategories[0].id : 0;
		selectedCurrency = defaultSelectedCurrency;
		currencySearch = '';
		amount = '';
		error = '';

		focusDialog();
	}

	$: if (!open && wasOpen) {
		wasOpen = false;
		pointerDownOutsideDialog = false;
		if (previouslyFocusedEl?.isConnected) {
			try {
				previouslyFocusedEl.focus();
			} catch {
				// Ignore focus errors if the previous element became unfocusable.
			}
		}
		previouslyFocusedEl = null;
	}

	async function handleSubmit() {
		if (submitting) return;
		error = '';

		if (!selectedCategoryId) {
			error = 'Please select a category.';
			return;
		}
		if (currencySearch.trim().length > 0 && filteredCurrencies.length === 0) {
			error = 'Please select a matching currency.';
			return;
		}
		if (!submitCurrency) {
			error = 'Please select a currency.';
			return;
		}
		if (!amount.trim()) {
			error = 'Please enter an amount.';
			return;
		}

		submitting = true;
		const result = await createExpense(token, selectedCategoryId, submitCurrency, amount.trim());
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

	function handleWindowPointerDown(e: PointerEvent) {
		if (!open || dialogEl === null) return;
		pointerDownOutsideDialog = !dialogEl.contains(e.target as Node);
	}

	function handleWindowPointerUp(e: PointerEvent) {
		if (!open || dialogEl === null) return;
		const pointerUpOutsideDialog = !dialogEl.contains(e.target as Node);
		if (pointerDownOutsideDialog && pointerUpOutsideDialog) onClose();
		pointerDownOutsideDialog = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!open) return;
		if (e.key === 'Escape') onClose();
	}
</script>

<svelte:window
	on:keydown={handleKeydown}
	on:pointerdown={handleWindowPointerDown}
	on:pointerup={handleWindowPointerUp}
/>

{#if open}
	<div class="backdrop">
		<div
			class="modal"
			role="dialog"
			aria-modal="true"
			aria-labelledby="add-expense-title"
			tabindex="-1"
			bind:this={dialogEl}
		>
			<h2 class="modal-title" id="add-expense-title">Add expense</h2>

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

			<label class="field-label" for="expense-currency-search">Currency</label>
			<input
				id="expense-currency-search"
				class="field-input"
				type="text"
				placeholder="Search currencies…"
				bind:value={currencySearch}
				disabled={submitting}
				autocomplete="off"
				spellcheck="false"
			/>
			{#if filteredCurrencies.length === 0}
				<p class="hint">No currencies match.</p>
			{:else}
				<select
					id="expense-currency"
					aria-label="Currency selection"
					class="field-select"
					bind:value={selectedCurrency}
					disabled={submitting}
				>
					{#each filteredCurrencies as code}
						<option value={code}>{code}</option>
					{/each}
				</select>
			{/if}

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
