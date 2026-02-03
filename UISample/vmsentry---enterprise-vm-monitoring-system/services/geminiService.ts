
import { GoogleGenAI } from "@google/genai";

// Initialize Gemini
const getAI = () => new GoogleGenAI({ apiKey: process.env.API_KEY || '' });

/**
 * Analyzes monitoring logs/alerts to suggest fixes
 */
export async function analyzeAlerts(alerts: string[], lang: string) {
  const ai = getAI();
  const prompt = `You are a VM systems expert. Analyze the following alerts and provide a concise summary of the root cause and recommended actions in ${lang === 'zh' ? 'Chinese' : lang === 'jp' ? 'Japanese' : 'English'}.
  Alerts: ${alerts.join('\n')}`;

  try {
    const response = await ai.models.generateContent({
      model: 'gemini-3-flash-preview',
      contents: prompt,
    });
    return response.text;
  } catch (error) {
    console.error("Gemini Error:", error);
    return "Error generating analysis.";
  }
}

/**
 * Edits an image based on a prompt (e.g., highlighting hardware issues in a data center photo)
 */
export async function editImageWithAI(base64Image: string, prompt: string) {
  const ai = getAI();
  try {
    const response = await ai.models.generateContent({
      model: 'gemini-2.5-flash-image',
      contents: {
        parts: [
          {
            inlineData: {
              data: base64Image.split(',')[1],
              mimeType: 'image/png',
            },
          },
          {
            text: prompt,
          },
        ],
      },
    });

    for (const part of response.candidates?.[0]?.content?.parts || []) {
      if (part.inlineData) {
        return `data:image/png;base64,${part.inlineData.data}`;
      }
    }
    return null;
  } catch (error) {
    console.error("Gemini Image Error:", error);
    throw error;
  }
}
