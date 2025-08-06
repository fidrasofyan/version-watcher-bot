import { kysely } from '../src/database';
import { notifyUsers } from '../src/job';

console.log('Notifying users...');
await notifyUsers();
console.log('Done!');

await kysely.destroy();
