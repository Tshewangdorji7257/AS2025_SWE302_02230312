import { NextResponse } from 'next/server';

// Helper function to retry API calls with exponential backoff
async function fetchWithRetry(url: string, maxRetries = 3): Promise<Response> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(url, {
        signal: AbortSignal.timeout(5000), // 5 second timeout
      });
      if (response.ok) return response;
      // If not ok but not last retry, wait and retry
      if (i < maxRetries - 1) {
        await new Promise(resolve => setTimeout(resolve, Math.pow(2, i) * 100)); // 100ms, 200ms, 400ms
        continue;
      }
      return response;
    } catch (error) {
      // If last retry, throw the error
      if (i === maxRetries - 1) throw error;
      // Otherwise wait and retry
      await new Promise(resolve => setTimeout(resolve, Math.pow(2, i) * 100));
    }
  }
  throw new Error('Max retries exceeded');
}

// GET /api/dogs - List all breeds
export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url);
    const breed = searchParams.get('breed');
    const count = searchParams.get('count') || '1';

    let apiUrl: string;

    if (breed) {
      apiUrl = `https://dog.ceo/api/breed/${breed}/images/random/${count}`;
    } else {
      apiUrl = `https://dog.ceo/api/breeds/image/random`;
    }

    const response = await fetchWithRetry(apiUrl);
    const data = await response.json();

    return NextResponse.json(data);
  } catch {
    return NextResponse.json(
      { error: 'Failed to fetch dog images' },
      { status: 500 }
    );
  }
}
