<script>
	import { goto } from '$app/navigation';

	let name = '';
	let password = '';
	let error = '';

	async function handleSubmit() {
		try {
			const response = await fetch('http://localhost:8080/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/x-www-form-urlencoded'
				},
				body: new URLSearchParams({
					name,
					password
				}).toString()
			});

			if (response.ok) {
				goto('/home');
			} else {
				const data = await response.json();
				error = data.message || 'Login failed';
			}
		} catch (err) {
			error = 'An error occurred during login';
		}
	}
</script>

<div class="flex flex-col items-center justify-center min-h-screen bg-gradient-to-br from-purple-500 via-pink-500 to-orange-500">
<form on:submit|preventDefault={handleSubmit} class="bg-white/95 p-8 rounded-xl shadow-[0_20px_50px_rgba(0,0,0,0.3)] w-96 backdrop-blur-sm">
	<h2 class="text-3xl font-bold mb-6 text-gray-800 text-center">Sign In</h2>
	
	{#if error}
		<div class="bg-red-100 text-red-700 p-3 rounded-lg mb-4 border border-red-200">
			{error}
		</div>
	{/if}
	
	<div class="mb-4">
		<input
			type="text"
			bind:value={name}
			placeholder="Username"
			class="w-full p-3 rounded-lg bg-gray-50 text-gray-800 placeholder-gray-500 border border-gray-200 focus:outline-none focus:border-purple-500 focus:ring-2 focus:ring-purple-200 transition duration-200"
			required
		/>
	</div>
	
	<div class="mb-6">
		<input
			type="password"
			bind:value={password}
			placeholder="Password"
			class="w-full p-3 rounded-lg bg-gray-50 text-gray-800 placeholder-gray-500 border border-gray-200 focus:outline-none focus:border-purple-500 focus:ring-2 focus:ring-purple-200 transition duration-200"
			required
		/>
	</div>
	
	<button
		type="submit"
		class="w-full bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 text-white font-semibold py-3 px-6 rounded-lg transition duration-300 shadow-md hover:shadow-lg transform hover:-translate-y-0.5"
	>
		Sign In
	</button>
	
	<div class="mt-4 text-center">
		<a href="/registration" class="text-gray-600 hover:text-purple-700 transition duration-200">Don't have an account? Sign up</a>
	</div>
</form>
</div>

