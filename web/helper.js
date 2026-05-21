export function formatDuration(ns) {
    if (!ns || ns <= 0) return 'N/A';
    if (ns >= 1000000000) {
        return (ns / 1000000000).toFixed(1) + 's';
    } else if (ns >= 1000000) {
        return (ns / 1000000).toFixed(1) + 'ms';
    } else if (ns >= 1000) {
        return (ns / 1000).toFixed(0) + 'us';
    } else {
        return Math.round(ns) + 'ns';
    }
}

export function parseTimeZone(date) {
    return new Date(date.getTime() - date.getTimezoneOffset() * 60000);
}

export function getDateDiff(date1, date2) {
    const diff = Math.abs(date2.getTime() - date1.getTime());
    const hours = Math.floor(diff / 3600000);
    const minutes = Math.floor((diff % 3600000) / 60000);
    const seconds = Math.floor((diff % 60000) / 1000);
    const pad = (n) => n.toString().padStart(2, '0');
    return { hours, minutes, seconds, toString: () => `${pad(hours)}:${pad(minutes)}:${pad(seconds)}` };
}

export function isValidLatinString(str) {
    return /^[\p{L}\p{N}!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`\s]+$/u.test(str);
}

export default {
    formatDuration,
    parseTimeZone,
    getDateDiff,
    isValidLatinString,
};