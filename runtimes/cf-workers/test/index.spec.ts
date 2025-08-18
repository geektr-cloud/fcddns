import { vi, describe, it, expect, beforeEach } from 'vitest';
import worker from '../src/index';
import { createExecutionContext, waitOnExecutionContext } from 'cloudflare:test';
import jwt from 'jsonwebtoken';

const mockDnsRecords = {
	list: vi.fn(),
	create: vi.fn(),
	update: vi.fn(),
};

const mockZones = {
	list: vi.fn(),
};

vi.mock('cloudflare', () => {
	const MockCloudflare = vi.fn(() => ({
		dns: {
			records: mockDnsRecords,
		},
		zones: mockZones,
	}));
	return {
		default: MockCloudflare,
	};
});

const env = { CLOUDFLARE_API_TOKEN: 'test-token', JWT_SECRET: 'test-secret' };
const host = 'test';
const domain = 'example.com';
const fqdn = `${host}.${domain}`;
const ip = '192.168.1.1';
const zoneId = 'test-zone-id';

describe('DDNS worker', () => {
	beforeEach(() => {
		vi.resetAllMocks();
	});

	const createRequest = (path: string, headers = {}) => {
		return new Request(`http://localhost${path}`, { headers });
	};

	const createJwt = (payload: object) => {
		return jwt.sign(payload, env.JWT_SECRET);
	};

	it('should return 404 for non-ddns paths', async () => {
		const request = createRequest('/not-ddns');
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(404);
		expect(await response.text()).toBe('Not found');
	});

	it('should return 400 if jwt is missing', async () => {
		const request = createRequest('/ddns/v2');
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(400);
		expect(await response.text()).toContain('Missing required parameters');
	});

	it('should return 400 if jwt is invalid', async () => {
		const request = createRequest('/ddns/v2/invalid-jwt');
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(400);
		expect(await response.text()).toBe('Invalid JWT');
	});

	it('should return 404 if zone is not found', async () => {
		const token = createJwt({ host, domain });
		mockZones.list.mockResolvedValue({ result: [] });
		const request = createRequest(`/ddns/v2/${token}/${ip}`);
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(404);
		expect(await response.text()).toBe(`Could not find zone for domain ${domain}`);
	});

	it('should create a new A record if none exists', async () => {
		const token = createJwt({ host, domain });
		mockZones.list.mockResolvedValue({ result: [{ id: zoneId, name: domain }] });
		mockDnsRecords.list.mockResolvedValue({ result: [] });
		const request = createRequest(`/ddns/v2/${token}/${ip}`);
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(201);
		expect(await response.text()).toBe(`Successfully created A record for ${fqdn} with IP ${ip}`);
		expect(mockDnsRecords.create).toHaveBeenCalledWith({
			zone_id: zoneId,
			type: 'A',
			name: fqdn,
			content: ip,
			ttl: 1,
			proxied: false,
		});
	});

	it('should update an existing A record if ip is different', async () => {
		const token = createJwt({ host, domain });
		const oldIp = '192.168.1.2';
		const recordId = 'test-record-id';
		mockZones.list.mockResolvedValue({ result: [{ id: zoneId, name: domain }] });
		mockDnsRecords.list.mockResolvedValue({ result: [{ id: recordId, content: oldIp, ttl: 120, proxied: true }] });
		const request = createRequest(`/ddns/v2/${token}/${ip}`);
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(200);
		expect(await response.text()).toBe(`Successfully updated IP for ${fqdn} to ${ip}`);
		expect(mockDnsRecords.update).toHaveBeenCalledWith(recordId, {
			zone_id: zoneId,
			type: 'A',
			name: fqdn,
			content: ip,
			ttl: 120,
			proxied: true,
		});
	});

	it('should return 200 if ip is already up to date', async () => {
		const token = createJwt({ host, domain });
		const recordId = 'test-record-id';
		mockZones.list.mockResolvedValue({ result: [{ id: zoneId, name: domain }] });
		mockDnsRecords.list.mockResolvedValue({ result: [{ id: recordId, content: ip, ttl: 120, proxied: true }] });
		const request = createRequest(`/ddns/v2/${token}/${ip}`);
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(200);
		expect(await response.text()).toBe(`IP for ${fqdn} is already up to date.`);
		expect(mockDnsRecords.update).not.toHaveBeenCalled();
	});

	it('should return 400 if multiple A records exist', async () => {
		const token = createJwt({ host, domain });
		mockZones.list.mockResolvedValue({ result: [{ id: zoneId, name: domain }] });
		mockDnsRecords.list.mockResolvedValue({ result: [{}, {}] });
		const request = createRequest(`/ddns/v2/${token}/${ip}`);
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(400);
		expect(await response.text()).toBe(`Multiple A records found for ${fqdn}, please manage them manually.`);
	});

	it('should use cf-connecting-ip header if ip is not in path', async () => {
		const token = createJwt({ host, domain });
		mockZones.list.mockResolvedValue({ result: [{ id: zoneId, name: domain }] });
		mockDnsRecords.list.mockResolvedValue({ result: [] });
		const request = createRequest(`/ddns/v2/${token}`, { 'cf-connecting-ip': ip });
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(201);
		expect(mockDnsRecords.create).toHaveBeenCalledWith(expect.objectContaining({ content: ip }));
	});

	it('should return 500 on unexpected errors', async () => {
		const token = createJwt({ host, domain });
		mockZones.list.mockRejectedValue(new Error('API error'));
		const request = createRequest(`/ddns/v2/${token}/${ip}`);
		const ctx = createExecutionContext();
		const response = await worker.fetch(request as any, env, ctx);
		await waitOnExecutionContext(ctx);
		expect(response.status).toBe(500);
		expect(await response.text()).toBe('An unexpected error occurred.');
	});
});
