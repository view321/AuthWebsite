<script>
	import Register from './Register.svelte';
	import Login from './Login.svelte';

	let page = 'register';
</script>

<nav>
	<button on:click={() => page = 'register'}>Register</button>
	<button on:click={() => page = 'login'}>Login</button>
</nav>

{#if page === 'register'}
	<Register />
{:else}
	<Login />
{/if}

<style>
	nav {
		margin-bottom: 1rem;
	}
	button {
		margin-right: 1rem;
	}
</style>