// image-upload.js: client-side WebP derivatives + MinIO presigned PUT

const IMAGE_SIZES = [
	{ size: '160w', width: 160, quality: 0.6 },
	{ size: '480w', width: 480, quality: 0.7 },
	{ size: '1200w', width: 1200, quality: 0.8 },
];

function adCsrfToken() {
	const el = document.querySelector('[hx-headers]');
	if (!el) {
		return '';
	}
	try {
		const headers = JSON.parse(el.getAttribute('hx-headers'));
		return headers['X-Csrf-Token'] || '';
	} catch (e) {
		return '';
	}
}

function showAdFormError(msg) {
	const el = document.getElementById('error');
	if (!el) {
		alert(msg);
		return;
	}
	el.textContent = msg;
	el.classList.remove('hidden');
}

function clearAdFormError() {
	const el = document.getElementById('error');
	if (!el) {
		return;
	}
	el.textContent = '';
	el.classList.add('hidden');
}

function setAdFormStatus(msg) {
	const el = document.getElementById('ad-form-status');
	const text = document.getElementById('ad-form-status-text');
	if (!el || !text) {
		return;
	}
	if (!msg) {
		text.textContent = '';
		el.classList.add('hidden');
		return;
	}
	text.textContent = msg;
	el.classList.remove('hidden');
}

function clearAdFormStatus() {
	setAdFormStatus('');
}

function setThumbnailProgress(container, pct) {
	let bar = container.querySelector('.upload-progress');
	if (!bar) {
		bar = document.createElement('div');
		bar.className = 'upload-progress absolute bottom-0 left-0 right-0 ' +
			'h-1.5 bg-zinc-200 rounded-b overflow-hidden';
		const fill = document.createElement('div');
		fill.className = 'upload-progress-fill h-full bg-blue-500';
		fill.style.width = '0%';
		bar.appendChild(fill);
		const thumb = container.querySelector('.relative');
		if (thumb) {
			thumb.appendChild(bar);
		}
	}
	const fill = bar.querySelector('.upload-progress-fill');
	if (fill) {
		fill.style.width = Math.max(0, Math.min(100, pct)) + '%';
	}
}

function setThumbnailError(container, failed) {
	container.style.outline = failed ? '2px solid #ef4444' : '';
}

async function loadImageBitmap(file) {
	if (typeof createImageBitmap === 'function') {
		return createImageBitmap(file);
	}
	return new Promise((resolve, reject) => {
		const url = URL.createObjectURL(file);
		const img = new Image();
		img.onload = () => {
			URL.revokeObjectURL(url);
			resolve(img);
		};
		img.onerror = () => {
			URL.revokeObjectURL(url);
			reject(new Error('Failed to decode image'));
		};
		img.src = url;
	});
}

function encodeWebP(source, maxWidth, quality) {
	const sw = source.width || source.naturalWidth;
	const sh = source.height || source.naturalHeight;
	const scale = sw > maxWidth ? maxWidth / sw : 1;
	const w = Math.max(1, Math.round(sw * scale));
	const h = Math.max(1, Math.round(sh * scale));
	const canvas = document.createElement('canvas');
	canvas.width = w;
	canvas.height = h;
	const ctx = canvas.getContext('2d');
	ctx.drawImage(source, 0, 0, w, h);
	return new Promise((resolve, reject) => {
		canvas.toBlob((blob) => {
			if (!blob) {
				reject(new Error('WebP encode failed'));
				return;
			}
			resolve(blob);
		}, 'image/webp', quality);
	});
}

async function prepareDerivatives(file, onProgress) {
	const bitmap = await loadImageBitmap(file);
	try {
		const out = {};
		const total = IMAGE_SIZES.length;
		for (let i = 0; i < total; i++) {
			const spec = IMAGE_SIZES[i];
			out[spec.size] = await encodeWebP(
				bitmap, spec.width, spec.quality);
			if (onProgress) {
				onProgress((i + 1) / total);
			}
		}
		return out;
	} finally {
		if (bitmap.close) {
			bitmap.close();
		}
	}
}

function putWithProgress(url, blob, onProgress) {
	return new Promise((resolve, reject) => {
		const xhr = new XMLHttpRequest();
		xhr.open('PUT', url);
		xhr.setRequestHeader('Content-Type', 'image/webp');
		xhr.upload.onprogress = (e) => {
			if (e.lengthComputable && onProgress) {
				onProgress(e.loaded / e.total);
			}
		};
		xhr.onload = () => {
			if (xhr.status >= 200 && xhr.status < 300) {
				resolve();
			} else {
				reject(new Error('Upload failed: ' + xhr.status));
			}
		};
		xhr.onerror = () => reject(new Error('Network error during upload'));
		xhr.send(blob);
	});
}

async function saveAdForm(form) {
	const url = form.getAttribute('data-ad-post-url');
	const body = new URLSearchParams(new FormData(form));
	body.delete('images');
	const res = await fetch(url, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
			'Accept': 'application/json',
			'X-Ad-Upload': '1',
			'X-Csrf-Token': adCsrfToken(),
		},
		body: body.toString(),
		credentials: 'same-origin',
	});
	if (!res.ok) {
		const data = await res.json().catch(() => ({}));
		throw new Error(data.error || 'Failed to save ad');
	}
	return res.json();
}

async function requestPresigns(adId, count) {
	const res = await fetch('/auth/ad/' + adId + '/presign-images', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			'Accept': 'application/json',
			'X-Csrf-Token': adCsrfToken(),
		},
		body: JSON.stringify({ count: count }),
		credentials: 'same-origin',
	});
	if (!res.ok) {
		const data = await res.json().catch(() => ({}));
		throw new Error(data.error || 'Failed to prepare uploads');
	}
	return res.json();
}

async function confirmImages(adId, imageCount) {
	const res = await fetch('/auth/ad/' + adId + '/confirm-images', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			'Accept': 'application/json',
			'X-Csrf-Token': adCsrfToken(),
		},
		body: JSON.stringify({ imageCount: imageCount }),
		credentials: 'same-origin',
	});
	if (!res.ok) {
		const data = await res.json().catch(() => ({}));
		throw new Error(data.error || 'Failed to confirm images');
	}
	return res.json();
}

async function uploadPreparedImages(adId, prepared) {
	const presign = await requestPresigns(adId, prepared.length);
	const byKey = {};
	presign.uploads.forEach((u) => {
		byKey[u.index + ':' + u.size] = u.putUrl;
	});

	for (let i = 0; i < prepared.length; i++) {
		const item = prepared[i];
		const index = presign.startIndex + i;
		const sizes = IMAGE_SIZES.map((s) => s.size);
		let done = 0;
		setThumbnailProgress(item.container, 0);
		for (const size of sizes) {
			const putUrl = byKey[index + ':' + size];
			if (!putUrl) {
				throw new Error('Missing upload URL for image ' + index);
			}
			await putWithProgress(putUrl, item.blobs[size], (p) => {
				const pct = ((done + p) / sizes.length) * 100;
				setThumbnailProgress(item.container, pct);
			});
			done += 1;
			setThumbnailProgress(item.container, (done / sizes.length) * 100);
		}
		setThumbnailError(item.container, false);
	}
	return presign.startIndex + prepared.length - 1;
}

function setSubmitBusy(submitBtn, busy) {
	if (!submitBtn) {
		return;
	}
	submitBtn.disabled = busy;
	submitBtn.classList.toggle('opacity-50', busy);
	submitBtn.classList.toggle('cursor-not-allowed', busy);
}

async function submitAdForm(event) {
	event.preventDefault();
	const form = event.target;
	if (form.dataset.uploading === '1') {
		return false;
	}
	clearAdFormError();

	const preview = document.getElementById('image-preview');
	const thumbs = preview
		? Array.from(preview.children).filter((el) => el.fileReference)
		: [];

	form.dataset.uploading = '1';
	const submitBtn = form.querySelector('button[type="submit"]');
	setSubmitBusy(submitBtn, true);

	try {
		setAdFormStatus('Saving ad...');
		const saved = await saveAdForm(form);
		const adId = saved.adId;
		let imageCount = saved.imageCount || 0;

		if (thumbs.length > 0) {
			const prepared = [];
			for (let i = 0; i < thumbs.length; i++) {
				const container = thumbs[i];
				const label = thumbs.length === 1
					? 'Processing image...'
					: 'Processing image ' + (i + 1) +
						' of ' + thumbs.length + '...';
				setAdFormStatus(label);
				setThumbnailError(container, false);
				setThumbnailProgress(container, 0);
				const blobs = await prepareDerivatives(
					container.fileReference,
					(p) => setThumbnailProgress(container, p * 100),
				);
				prepared.push({ container: container, blobs: blobs });
			}
			try {
				setAdFormStatus('Uploading images...');
				imageCount = await uploadPreparedImages(adId, prepared);
				setAdFormStatus('Finishing...');
				await confirmImages(adId, imageCount);
			} catch (uploadErr) {
				thumbs.forEach((c) => setThumbnailError(c, true));
				throw uploadErr;
			}
		}

		setAdFormStatus('Redirecting...');
		window.location.href = '/ad/' + adId;
	} catch (err) {
		console.error(err);
		clearAdFormStatus();
		showAdFormError(err.message || 'Something went wrong');
		form.dataset.uploading = '0';
		setSubmitBusy(submitBtn, false);
	}
	return false;
}
