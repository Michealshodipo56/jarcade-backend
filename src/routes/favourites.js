const express = require('express');
const { body, validationResult } = require('express-validator');
const db = require('../db');
const { authMiddleware } = require('../middleware/auth');

const router = express.Router();

router.use(authMiddleware);

function toFavourite(row) {
  return {
    id: row.id,
    name: row.game_name,
    img: row.game_img || '',
    onclick: row.play_onclick || '',
    createdAt: row.created_at,
  };
}

router.get('/', (req, res) => {
  const rows = db
    .prepare(
      `SELECT id, game_name, game_img, play_onclick, created_at
       FROM favourites WHERE user_id = ? ORDER BY created_at DESC`
    )
    .all(req.user.sub);
  res.json({ favourites: rows.map(toFavourite) });
});

router.post(
  '/',
  [
    body('name').trim().notEmpty().withMessage('Game name is required.'),
    body('img').optional().isString(),
    body('onclick').optional().isString(),
  ],
  (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      return res.status(400).json({ error: errors.array()[0].msg });
    }

    const { name, img = '', onclick = '' } = req.body;

    try {
      db.prepare(
        `INSERT INTO favourites (user_id, game_name, game_img, play_onclick)
         VALUES (?, ?, ?, ?)
         ON CONFLICT(user_id, game_name) DO UPDATE SET
           game_img = excluded.game_img,
           play_onclick = excluded.play_onclick`
      ).run(req.user.sub, name, img, onclick);
    } catch (err) {
      console.error(err);
      return res.status(500).json({ error: 'Could not save favourite.' });
    }

    const row = db
      .prepare(
        `SELECT id, game_name, game_img, play_onclick, created_at
         FROM favourites WHERE user_id = ? AND game_name = ?`
      )
      .get(req.user.sub, name);

    res.status(201).json({ favourite: toFavourite(row) });
  }
);

router.delete('/:name', (req, res) => {
  const name = decodeURIComponent(req.params.name);
  const result = db
    .prepare('DELETE FROM favourites WHERE user_id = ? AND game_name = ?')
    .run(req.user.sub, name);

  if (result.changes === 0) {
    return res.status(404).json({ error: 'Favourite not found.' });
  }
  res.json({ ok: true });
});

module.exports = router;
