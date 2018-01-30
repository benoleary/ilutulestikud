import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { environment } from '../environments/environment';
import { Player } from './models/player.model'

@Injectable()
export class IlutulestikudService {
  httpClient: HttpClient;
  uriRoot: string;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
    this.uriRoot = environment.restRoot;
  }

  registeredPlayers(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "lobby/registered-players")
  }

  availableColors(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "lobby/available-colors")
  }

  newPlayer(newPlayerName: string): Observable<any> {
    return this.httpClient.post(this.uriRoot + "lobby/new-player", {"Name": newPlayerName})
  }

  updatePlayer(playerOverride: Player): Observable<any> {
    return this.httpClient.post(this.uriRoot + "lobby/update-player", playerOverride)
  }
}
