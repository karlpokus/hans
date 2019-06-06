#!/usr/bin/env node

const delay = 5000;
const secret = process.env.SECRET_MSG || "unknown";
const loop = () => {
	console.log(`The secret is ${ secret }`);
}

setInterval(loop, delay)
