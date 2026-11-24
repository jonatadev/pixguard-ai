import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import {
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
}
