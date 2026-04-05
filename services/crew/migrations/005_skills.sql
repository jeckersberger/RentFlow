-- Add skills column to crew_members as a TEXT array.
ALTER TABLE crew_members ADD COLUMN IF NOT EXISTS skills TEXT[] DEFAULT '{}';

-- GIN index for fast array-contains lookups on skills.
CREATE INDEX IF NOT EXISTS idx_crew_members_skills ON crew_members USING GIN (skills);
