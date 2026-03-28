import api from './api';
import type { CrewMember, CrewAssignment, CrewQualification } from '@/types/crew';

export async function listCrewMembers(): Promise<CrewMember[]> {
  return api.get('/api/v1/crew') as unknown as CrewMember[];
}

export async function getCrewMember(id: string): Promise<CrewMember> {
  return api.get(`/api/v1/crew/${id}`) as unknown as CrewMember;
}

export async function createCrewMember(body: Partial<CrewMember>): Promise<CrewMember> {
  return api.post('/api/v1/crew', body) as unknown as CrewMember;
}

export async function updateCrewMember(id: string, body: Partial<CrewMember>): Promise<CrewMember> {
  return api.put(`/api/v1/crew/${id}`, body) as unknown as CrewMember;
}

export async function deleteCrewMember(id: string): Promise<void> {
  await api.delete(`/api/v1/crew/${id}`);
}

export async function listAssignments(memberId: string): Promise<CrewAssignment[]> {
  return api.get(`/api/v1/crew/${memberId}/assignments`) as unknown as CrewAssignment[];
}

export async function createAssignment(memberId: string, body: Partial<CrewAssignment>): Promise<CrewAssignment> {
  return api.post(`/api/v1/crew/${memberId}/assignments`, body) as unknown as CrewAssignment;
}

export async function listAllAssignments(params?: { project_id?: string }): Promise<CrewAssignment[]> {
  return api.get('/api/v1/crew-assignments', { params }) as unknown as CrewAssignment[];
}

export async function listQualifications(memberId: string): Promise<CrewQualification[]> {
  return api.get(`/api/v1/crew/${memberId}/qualifications`) as unknown as CrewQualification[];
}

export async function createQualification(memberId: string, body: Partial<CrewQualification>): Promise<CrewQualification> {
  return api.post(`/api/v1/crew/${memberId}/qualifications`, body) as unknown as CrewQualification;
}
