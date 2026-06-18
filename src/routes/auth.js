const express = require('express');
const bcrypt = require('bcryptjs');
const { body, validationResult } = require('express-validator');
const db = require('../db');
const { signToken, authMiddleware } = require('../middleware/auth');

const router = express.Router();
const SALT_ROUNDS = 12;

function publicUser(row) {
  return {
    id: row.id,
    username: row.username,
    email: row.email,
    initials: row.username.slice(0, 2).toUpperCase(),
    createdAt: row.created_at,
  };
}

function handleValidation(req, res) {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    res.status(400).json({ error: errors.array()[0].msg, errors: errors.array() });
    return true;
  }
  return false;
}

router.post(
  '/register',
  [
    body('username')
      .trim()
      .isLength({ min: 3, max: 24 })
      .withMessage('Username must be 3–24 characters.')
      .matches(/^[a-zA-Z0-9_]+$/)
      .withMessage('Username may only contain letters, numbers, and underscores.'),
    body('email').trim().isEmail().withMessage('Enter a valid email address.').normalizeEmail(),
    body('password')
      .isLength({ min: 8 })
      .withMessage('Password must be at least 8 characters.'),
    body('confirmPassword').custom((value, { req }) => {
      if (value !== req.body.password) throw new Error('Passwords do not match.');
      return true;
    }),
  ],
  async (req, res) => {
    if (handleValidation(req, res)) return;

    const { username, email, password } = req.body;

    const existing = db
      .prepare('SELECT id FROM users WHERE username = ? COLLATE NOCASE OR email = ? COLLATE NOCASE')
      .get(username, email);

    if (existing) {
      return res.status(409).json({ error: 'Username or email is already registered.' });
    }

    const passwordHash = await bcrypt.hash(password, SALT_ROUNDS);

    let user;
    try {
      const result = db
        .prepare('INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)')
        .run(username, email, passwordHash);
      user = db.prepare('SELECT * FROM users WHERE id = ?').get(result.lastInsertRowid);
    } catch (err) {
      if (err.code === 'SQLITE_CONSTRAINT_UNIQUE') {
        return res.status(409).json({ error: 'Username or email is already registered.' });
      }
      throw err;
    }

    const token = signToken(user, false);
    res.status(201).json({ user: publicUser(user), token });
  }
);

router.post(
  '/login',
  [
    body('login').trim().notEmpty().withMessage('Enter your username or email.'),
    body('password').notEmpty().withMessage('Enter your password.'),
  ],
  async (req, res) => {
    if (handleValidation(req, res)) return;

    const { login, password, remember } = req.body;

    const user = db
      .prepare(
        `SELECT * FROM users
         WHERE username = ? COLLATE NOCASE OR email = ? COLLATE NOCASE`
      )
      .get(login, login);

    if (!user) {
      return res.status(401).json({ error: 'Invalid username/email or password.' });
    }

    const valid = await bcrypt.compare(password, user.password_hash);
    if (!valid) {
      return res.status(401).json({ error: 'Invalid username/email or password.' });
    }

    const token = signToken(user, !!remember);
    res.json({ user: publicUser(user), token });
  }
);

router.get('/me', authMiddleware, (req, res) => {
  const user = db.prepare('SELECT * FROM users WHERE id = ?').get(req.user.sub);
  if (!user) {
    return res.status(401).json({ error: 'User not found.' });
  }
  res.json({ user: publicUser(user) });
});

router.post('/logout', (_req, res) => {
  res.json({ ok: true });
});

module.exports = router;
