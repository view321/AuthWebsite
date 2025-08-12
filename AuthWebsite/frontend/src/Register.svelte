<script>
	let username = '';
	let password = '';
	let message = '';

	async function register() {
		const res = await fetch('http://localhost:8080/api/register', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username, password })
		});
		const data = await res.json();
		message = data.message;
	}
</script>

<h2>Register</h2>
<input placeholder="Username" bind:value={username}>
<input type="password" placeholder="Password" bind:value={password}>
<button on:click={register}>Register</button>

<p>{message}</p>