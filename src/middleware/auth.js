const jwt = require('jsonwebtoken');

const SECRET = process.env.JWT_SECRET || 'dev-only-change-me';
const DEFAULT_EXPIRES = process.env.JWT_EXPIRES_IN || '7d';
const REMEMBER_EXPIRES = process.env.JWT_REMEMBER_EXPIRES_IN || '30d';

function signToken(user, remember = false) {
  return jwt.sign(
    { sub: user.id, username: user.username, email: user.email },
    SECRET,
    { expiresIn: remember ? REMEMBER_EXPIRES : DEFAULT_EXPIRES }
  );
}

function verifyToken(token) {
  return jwt.verify(token, SECRET);
}

function authMiddleware(req, res, next) {
  const header = req.headers.authorization || '';
  const token = header.startsWith('Bearer ') ? header.slice(7) : null;

  if (!token) {
    return res.status(401).json({ error: 'Authentication required.' });
  }

  try {
    req.user = verifyToken(token);
    next();
  } catch {
    return res.status(401).json({ error: 'Invalid or expired session.' });
  }
}

module.exports = { signToken, verifyToken, authMiddleware };
