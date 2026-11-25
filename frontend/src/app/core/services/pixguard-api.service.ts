import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import {
  AnalysisEvent,
  AnalysisResult,
  AnalystFeedbackRecord,
  AnalystFeedbackRequest,
  CaseDetailResponse,
  CaseListResponse,
  GuardrailPreviewResponse,
  HealthResponse,
  ReadyResponse,
} from '../models/pix.models';

@Injectable({
  providedIn: 'root',
})
export class PixguardApiService {
  private readonly http = inject(HttpClient);

  getHealth(): Observable<HealthResponse> {
    return this.http.get<HealthResponse>('/healthz');
  }

  getReady(): Observable<ReadyResponse> {
    return this.http.get<ReadyResponse>('/readyz');
  }

  listCases(risk = ''): Observable<CaseListResponse> {
    let params = new HttpParams();
    if (risk && risk !== 'TODOS') {
      params = params.set('risk', risk);
    }
    return this.http.get<CaseListResponse>('/api/v1/cases', { params });
  }

  getCaseDetail(id: string): Observable<CaseDetailResponse> {
    return this.http.get<CaseDetailResponse>(`/api/v1/cases/${encodeURIComponent(id)}`);
  }

  previewGuardrail(text: string): Observable<GuardrailPreviewResponse> {
    return this.http.post<GuardrailPreviewResponse>('/api/v1/guardrail/preview', { text });
  }

  async analyzeCaseStream(
    id: string,
    onStage: (event: AnalysisEvent) => void
  ): Promise<AnalysisResult> {
    const response = await fetch(`/api/v1/cases/${encodeURIComponent(id)}/analyze`, {
      method: 'POST',
      headers: {
        Accept: 'text/event-stream',
      },
    });

    if (!response.ok || !response.body) {
      throw new Error(`Falha ao analisar caso (${response.status})`);
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let finalResult: AnalysisResult | null = null;

    while (true) {
      const { value, done } = await reader.read();
      if (done) {
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const blocks = buffer.split('\n\n');
      buffer = blocks.pop() ?? '';

      for (const block of blocks) {
        const lines = block.split('\n');
        let eventName = 'message';
        let dataLine = '';
        for (const line of lines) {
          if (line.startsWith('event:')) {
            eventName = line.slice(6).trim();
          } else if (line.startsWith('data:')) {
            dataLine += line.slice(5).trim();
          }
        }
        if (!dataLine) {
          continue;
        }
        const parsed = JSON.parse(dataLine);
        if (eventName === 'stage') {
          onStage(parsed as AnalysisEvent);
        } else if (eventName === 'complete') {
          finalResult = parsed as AnalysisResult;
        }
      }
    }

    if (!finalResult) {
      throw new Error('Stream SSE encerrado sem evento complete');
    }
    return finalResult;
  }

  submitFeedback(id: string, payload: AnalystFeedbackRequest): Observable<AnalystFeedbackRecord> {
    return this.http.post<AnalystFeedbackRecord>(
      `/api/v1/cases/${encodeURIComponent(id)}/feedback`,
      payload
    );
  }
}
