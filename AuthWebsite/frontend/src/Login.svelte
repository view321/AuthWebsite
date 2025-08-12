<script>
	let username = '';
	let password = '';
	let message = '';

	async function login() {
		const res = await fetch('http://localhost:8080/api/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username, password })
		});
		const data = await res.json();
		message = data.message;
	}
</script>

<h2>Login</h2>
<input placeholder="Username" bind:value={username}>
<input type="password" placeholder="Password" bind:value={password}>
<button on:click={login}>Login</button>

<p>{message}</p>