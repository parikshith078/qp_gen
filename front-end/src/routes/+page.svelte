<script lang="ts">
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<h1>Home</h1>
<pre>{JSON.stringify(data.user, null, 1)}</pre>
