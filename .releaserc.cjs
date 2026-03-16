/**
 * semantic-release config (JS so we can use functions + comments)
 *
 * Notes strategy:
 * - A single release-notes-generator produces notes used by both CHANGELOG.md and GitHub Release.
 * - Using two generators would cause semantic-release to concatenate their outputs, duplicating entries.
 */
'use strict';

const fs   = require('fs');
const yaml = require('js-yaml');

// Load valid scopes from the unified taxonomy in tags.yml.
// Both "both" and "scopes" entries are valid commit scopes.
const _tagDef     = yaml.load(fs.readFileSync(new URL('.github/tags.yml', `file://${__dirname}/`), 'utf8'));
const validScopes = new Set([
  ...(_tagDef.tags.both   || []),
  ...(_tagDef.tags.scopes || []),
]);

const OTHER_SECTION = '🧩 Other';

const SECTION_TITLES = {
  feat: '✨ Features',
  fix: '🐛 Fixes',
  perf: '⚡ Performance',
  test: '✅ Tests',
  build: '📦 Build',
  ci: '🤖 CI / CD',
  chore: '🧹 Chores',
  style: '💄 Style',
  refactor: '♻️ Refactors',  
  docs: '📝 Docs',
  post: '✉️ Posts',
};

// Order groups exactly as declared above, and always render "Other" last.
const GROUP_ORDER = Object.values(SECTION_TITLES);

const ALLOWED_TYPES_FOR_NOTES = new Set(Object.keys(SECTION_TITLES));

/**
 * Notes transform policy
 * - Skips merge commits entirely
 * - Skips commits with no subject (prevents empty bullets)
 * - Routes unknown/missing types into the "Other" group
 * - Returns a NEW object (immutable-safe)
 */
function baseTransform(commit) {
  // conventional-commits-parser sets commit.merge for merge commits
  if (commit.merge || /^Merge\b/i.test(commit.subject || '')) return;

  const subject = (commit.subject || '').trim();
  if (!subject) return;

  const scope = (commit.scope || '').trim();
  const flaggedSubject =
    scope && !validScopes.has(scope) ? `${subject} ⚠️ unknown-scope` : subject;

  const rawType = (commit.type || '').trim();
  const normalizedType =
    rawType && ALLOWED_TYPES_FOR_NOTES.has(rawType) ? rawType : 'other';

  return {
    ...commit,
    subject: flaggedSubject,
    type:
      normalizedType === 'other'
        ? OTHER_SECTION
        : SECTION_TITLES[normalizedType],
    shortHash: commit.hash?.slice(0, 7),
  };
}

function commitGroupsSort(a, b) {
  // Always last
  if (a.title === OTHER_SECTION && b.title !== OTHER_SECTION) return 1;
  if (a.title !== OTHER_SECTION && b.title === OTHER_SECTION) return -1;

  // Order by SECTION_TITLES placement
  const ai = GROUP_ORDER.indexOf(a.title);
  const bi = GROUP_ORDER.indexOf(b.title);

  // Known groups first, in declared order
  if (ai !== -1 && bi !== -1) return ai - bi;

  // If one is known and the other isn't, known wins
  if (ai !== -1 && bi === -1) return -1;
  if (ai === -1 && bi !== -1) return 1;

  // Fallback
  return a.title.localeCompare(b.title);
}

/**
 * We set mainTemplate explicitly because the preset defaults sometimes flatten output.
 * This forces section headers and deterministic ordering.
 */
const changelogMainTemplate = [
  '## 📦 Release {{version}}',
  '',
  '{{#each commitGroups}}',
  '### {{title}}',
  '',
  '{{#each commits}}',
  '{{> commit}}',
  '{{/each}}',
  '',
  '{{/each}}',
].join('\n');

/** Writer opts shared by both CHANGELOG.md and GitHub Release */
const changelogWriterOpts = {
  groupBy: 'type',
  commitGroupsSort,
  commitsSort: ['scope', 'subject'],
  transform: baseTransform,

  mainTemplate: changelogMainTemplate,

  // IMPORTANT: include newline so bullets don't run together
  commitPartial:
    '- {{#if scope}}**{{scope}}:** {{/if}}{{subject}} ({{shortHash}})\n',
};


module.exports = {
  branches: [
    'main',
    { name: 'canary', channel: 'canary', prerelease: true },
  ],
  tagFormat: 'v${version}',
  plugins: [
    // 1) Decide version bump based on commits
    [
      '@semantic-release/commit-analyzer',
      {
        preset: 'conventionalcommits',
        releaseRules: [
          { breaking: true, release: 'major' },
          { type: 'feat', release: 'minor' },
          { type: 'fix', release: 'patch' },
          { type: 'perf', release: 'patch' },

          // Everything else: no release bump
          { type: 'refactor', release: false },
          { type: 'docs', release: false },
          { type: 'chore', release: false },
          { type: 'test', release: false },
          { type: 'ci', release: false },
          { type: 'style', release: false },
          { type: 'build', release: false },
        ],
      },
    ],

    // 2) Generate release notes (used by both CHANGELOG.md and GitHub Release)
    [
      '@semantic-release/release-notes-generator',
      {
        preset: 'conventionalcommits',
        writerOpts: changelogWriterOpts,
      },
    ],

    // 3) Update CHANGELOG.md
    [
      '@semantic-release/changelog',
      {
        changelogFile: 'CHANGELOG.md',
        changelogTitle: '# 📦 Release History',
      },
    ],

    // 4) Publish GitHub Release
    '@semantic-release/github',

    // 5) Commit CHANGELOG.md back to the repo
    [
      '@semantic-release/git',
      {
        assets: ['CHANGELOG.md'],
        message: 'chore(release): ${nextRelease.version}',
      },
    ],
  ],
};
