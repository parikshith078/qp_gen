<script lang="ts">
	import { enhance } from '$app/forms';

	let { form } = $props();
</script>

<div class=" container mx-auto">
	<h1>Register</h1>
	{#if form?.error}
		<p>Error: {form.error}</p>
	{/if}
	<form class="flex flex-col gap-1" method="POST" use:enhance>
		<label>
			Email:
			<input type="email" required name="email" value={form?.payload.email ?? ''} id="email" />
		</label>
		<label>
			Password:
			<input
				type="password"
				required
				value={form?.payload.password ?? ''}
				name="password"
				id="password"
			/>
		</label>
		<button type="submit">sumbit</button>
	</form>
</div>
