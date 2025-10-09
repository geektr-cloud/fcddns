import Cloudflare from 'cloudflare';
import { jwtVerify } from 'jose';

interface JwtPayload {
	host: string;
	domain: string;
	slot?: number;
	ip?: string;
}

const slotMarkPattern = /\[fcddns-slot:\d+\]/;
const slotMarkF = (d: number) => `[fcddns-slot:${d}]`;

export default {
	async fetch(request, env, ctx): Promise<Response> {
		const { pathname } = new URL(request.url);

		if (pathname === '/myip') {
			return new Response(request.headers.get('cf-connecting-ip') as string, { status: 200 });
		}

		if (!pathname.startsWith('/ddns/v1')) {
			return new Response('Not found', { status: 404 });
		}

		// /ddns/v2/{jwt}[/ip]
		let [, , , token, ip] = pathname.split('/');
		if (!token) {
			return new Response('Missing required parameters: jwt must be provided and must be a valid JWT', { status: 400 });
		}

		let decoded: JwtPayload;
		try {
			const secret = new TextEncoder().encode(env.JWT_SECRET);
			decoded = await jwtVerify(token, secret, { algorithms: ['HS256'], currentDate: new Date(0) }).then(
				(res) => res.payload as unknown as JwtPayload
			);
		} catch (e) {
			return new Response('Invalid JWT', { status: 400 });
		}

		if (!decoded) {
			return new Response('Invalid JWT', { status: 400 });
		}

		const { host, domain, slot = 0 } = decoded;
		const slotMark = slotMarkF(slot);

		ip = ip || decoded.ip || (request.headers.get('cf-connecting-ip') as string);
		const deleteRecord = ip === '0.0.0.0';

		const cf = new Cloudflare({ apiToken: env.CLOUDFLARE_API_TOKEN });
		const fqdn = `${host}.${domain}`;

		try {
			const zone = await cf.zones.list({ name: domain }).then((res) => res.result.find((z) => z.name === domain));
			if (!zone) {
				return new Response(`Could not find zone for domain ${domain}`, { status: 404 });
			}
			const zoneId = zone.id;

			const recordsResponse = await cf.dns.records.list({ zone_id: zoneId, type: 'A', name: { exact: fqdn } });
			const records = recordsResponse.result;

			const nonManagedRecords = records.filter((record) => !slotMarkPattern.test(record.comment ?? ''));
			if (nonManagedRecords.length > 0) {
				return new Response(
					`There is ${nonManagedRecords.length} record(s) not managed by fcddns for ${fqdn}, please manage them manually.`,
					{ status: 400 }
				);
			}

			const record = records.find((record) => record.comment?.includes(slotMark));

			if (record === undefined) {
				if (deleteRecord) return new Response(`Record for ${fqdn} in slot ${slot} not found, so delete record is skipped`, { status: 200 });

				await cf.dns.records.create({
					zone_id: zoneId,
					type: 'A',
					name: fqdn,
					content: ip,
					ttl: 1,
					proxied: false,
					comment: slotMark,
				});
				return new Response(`Successfully created A record for ${fqdn} with IP ${ip} in slot ${slot}`, { status: 201 });
			}

			if (record.content === ip) {
				return new Response(`IP for ${fqdn} is already up to date.`, { status: 200 });
			}

			if (deleteRecord) {
				await cf.dns.records.delete(record.id, { zone_id: zoneId });
				return new Response(`Successfully deleted A record for ${fqdn} in slot ${slot}`, { status: 200 });
			}

			await cf.dns.records.update(record.id, {
				zone_id: zoneId,
				type: 'A',
				name: fqdn,
				content: ip,
				ttl: record.ttl,
				proxied: record.proxied,
				comment: slotMark,
			});
			return new Response(`Successfully updated IP for ${fqdn} to ${ip} in slot ${slot}`, { status: 200 });
		} catch (error) {
			console.error('DDNS update failed:', error);
			return new Response('An unexpected error occurred.', { status: 500 });
		}
	},
} satisfies ExportedHandler<Env>;
