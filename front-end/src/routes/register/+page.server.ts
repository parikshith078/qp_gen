import { api } from '$lib/apis/axiosConfig';
import { type ApiResponse } from '$lib/apis/types';
import { fail, redirect, type Actions } from '@sveltejs/kit';
import type { AxiosError } from 'axios';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ cookies }) => {
	const userid = cookies.get('userid');
	console.log(userid);
	const useremail = cookies.get('useremail');
	if (useremail && userid) {
		throw redirect(307, '/');
	}
};
export const actions = {
	default: async ({ request }) => {
		// Extract form data
		const data = await request.formData();
		const payload = {
			name: data.get('name'),
			email: data.get('email'),
			username: data.get('username'),
			password: data.get('password')
		};

		try {
			// Attempt to register the user
			await api.post<ApiResponse>('/register', payload);
		} catch (error) {
			// Handle API errors
			const axiosError = error as AxiosError<ApiResponse>;
			const statusCode = axiosError.response?.status ?? 400;
			const errMsg = axiosError.response?.data?.message ?? 'Registration failed. Please try again.';

			// Return the error and original form data to repopulate the form
			return fail(statusCode, {
				payload,
				error: errMsg
			});
		}
		// If we get here, registration was successful
		return redirect(303, '/login?registered=true');
	}
} satisfies Actions;
