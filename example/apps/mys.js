#!/usr/bin/env node

const name = 'mys'
const delay = 12000;
const loop = () => {
	console.log(`I am ${ name } and I log every ${ delay } msecs`);
}

setInterval(loop, delay)
