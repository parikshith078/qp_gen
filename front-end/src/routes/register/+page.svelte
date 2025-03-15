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
			Name:
			<input type="text" value={form?.payload.name ?? ''} name="name" id="name" />
		</label>
		<label>
			Email:
			<input type="email" name="email" value={form?.payload.email ?? ''} id="email" />
		</label>
		<label>
			Username:
			<input type="text" value={form?.payload.username ?? ''} name="username" id="username" />
		</label>
		<label>
			Password:
			<input type="password" value={form?.payload.password ?? ''} name="password" id="password" />
		</label>
		<button type="submit">sumbit</button>
	</form>
</div>
