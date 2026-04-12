import { describe, it, expect, vi } from 'vitest';
import express from 'express';
import request from 'supertest';
import { verifyToken } from '../src/middleware';

describe('OathMesh Express Middleware', () => {
    it('blocks request without Authorization header', async () => {
        const app = express();
        app.use(verifyToken({ audience: 'bench', trustedIssuers: ['http://issuer.local'] }));
        app.get('/', (req, res) => res.send('ok'));
        
        const res = await request(app).get('/');
        expect(res.status).toBe(401);
        expect(res.body.code).toBe('claim_missing:token');
    });

    it('blocks request with invalid Authorization form', async () => {
        const app = express();
        app.use(verifyToken({ audience: 'bench', trustedIssuers: ['http://issuer.local'] }));
        app.get('/', (req, res) => res.send('ok'));
        
        const res = await request(app).get('/').set('Authorization', 'Bearer fake');
        expect(res.status).toBe(401);
        expect(res.body.code).toBe('claim_missing:token');
    });

    it('blocks requests with malformed tokens', async () => {
        const app = express();
        app.use(verifyToken({ audience: 'bench', trustedIssuers: ['http://issuer.local'] }));
        app.get('/', (req, res) => res.send('ok'));
        
        const res = await request(app).get('/').set('Authorization', 'OathMesh badlyformed');
        expect(res.status).toBe(401);
        expect(res.body.code).toBe('verification_failed');
    });
});
