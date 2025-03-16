import { api } from '$lib/apis/axiosConfig';
import { type ApiResponse, type UserModel } from '$lib/apis/types';
import { fail, redirect, type Actions } from '@sveltejs/kit';
import type { AxiosError } from 'axios';
import type { PageServerLoad } from './$types';
import { extractTokensFromHeaders } from '$lib/utils/session';

export const load: PageServerLoad = ({ cookies }) => {
	const userid = cookies.get('userid');
	const useremail = cookies.get('useremail');
	if (useremail && userid) {
		throw redirect(307, '/');
	}
};

export const actions = {
	default: async ({ request, cookies }) => {
		// Extract form data
		const data = await request.formData();
		const payload = {
			email: data.get('email'),
			password: data.get('password')
		};

		try {
			// Attempt to login the user
			const { data, headers } = await api.post<ApiResponse<UserModel>>('/login', payload);

			// Get cookies from the response
			const { sessionToken, csrfToken } = extractTokensFromHeaders(headers);
			if (sessionToken && csrfToken) {
				cookies.set(sessionToken.name, sessionToken.value, {
					path: '/',
					expires: sessionToken.expires,
					httpOnly: sessionToken.httpOnly
				});
				cookies.set(csrfToken.name, csrfToken.value, {
					path: '/',
					expires: csrfToken.expires,
					httpOnly: csrfToken.httpOnly
				});
			}

			// Still set your application cookies
			if (data.data) {
				cookies.set('userid', data.data.id, { path: '/', httpOnly: false });
			} else {
				return fail(501, {
					payload,
					error: 'Login failed1, please try again'
				});
			}
		} catch (error) {
			// Handle API errors
			const axiosError = error as AxiosError<ApiResponse>;
			const statusCode = axiosError.response?.status ?? 400;
			const errorMessage = axiosError.response?.data?.message ?? 'Login failed. Please try again. ';

			// Return the error and original form data to repopulate the form
			return fail(statusCode, {
				payload,
				error: errorMessage
			});
		}
		// If we get here, login was successful
		return redirect(303, '/');
	}
} satisfies Actions;
