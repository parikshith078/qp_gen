interface Cookie {
	name: string;
	value: string;
	expires?: Date;
	httpOnly?: boolean;
}

interface TokensResult {
	csrfToken: Cookie | null;
	sessionToken: Cookie | null;
}

interface Cookie {
	name: string;
	value: string;
	expires?: Date;
	httpOnly?: boolean;
}

interface TokensResult {
	csrfToken: Cookie | null;
	sessionToken: Cookie | null;
}

export function extractTokensFromHeaders(headers: any): TokensResult {
	let csrfToken: Cookie | null = null;
	let sessionToken: Cookie | null = null;

	// Check if we have set-cookie headers
	const cookies = headers['set-cookie'];

	if (Array.isArray(cookies)) {
		cookies.forEach((cookie: string) => {
			// Extract the full name=value part first without splitting on '='
			const firstSemicolonIndex = cookie.indexOf(';');
			const nameValuePair =
				firstSemicolonIndex !== -1 ? cookie.substring(0, firstSemicolonIndex) : cookie;

			// Now split the name-value pair at the first '=' only
			const equalsIndex = nameValuePair.indexOf('=');
			if (equalsIndex === -1) return; // Skip if no '=' found

			const name = nameValuePair.substring(0, equalsIndex);
			// Get everything after the first '=' to preserve any '=' in the value itself
			const value = nameValuePair.substring(equalsIndex + 1);

			const cookieObj: Cookie = {
				name,
				value
			};

			// Extract other properties from the parts after the first semicolon
			if (firstSemicolonIndex !== -1) {
				const parts = cookie
					.substring(firstSemicolonIndex + 1)
					.split(';')
					.map((part) => part.trim());

				parts.forEach((part) => {
					if (part.toLowerCase().startsWith('expires=')) {
						cookieObj.expires = new Date(part.substring(8));
					} else if (part.toLowerCase() === 'httponly') {
						cookieObj.httpOnly = true;
					}
					// Could add more properties like Path, Secure, etc. as needed
				});
			}

			// Store the token in the appropriate variable
			if (name === 'csrf_token') {
				csrfToken = cookieObj;
			} else if (name === 'session_token') {
				sessionToken = cookieObj;
			}
		});
	}

	return {
		csrfToken,
		sessionToken
	};
}
