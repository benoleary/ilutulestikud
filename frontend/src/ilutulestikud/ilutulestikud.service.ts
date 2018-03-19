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
    return this.httpClient.get(this.uriRoot + "player/registered-players")
  }

  availableColors(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "player/available-colors")
  }

  newPlayer(newPlayerName: string): Observable<any> {
    return this.httpClient.post(
      this.uriRoot + "player/new-player",
      {
        "Name": newPlayerName
      })
  }

  updatePlayer(playerOverride: Player): Observable<any> {
    return this.httpClient.post(
      this.uriRoot + "player/update-player",
      playerOverride)
  }

  availableRulesets(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "game/available-rulesets")
  }

  gamesWithPlayer(playerIdentifier: string): Observable<any> {
    return this.httpClient.get(
      this.uriRoot + "game/all-games-with-player/" + encodeURIComponent(playerIdentifier))
  }

  newGame(newGameName: string, rulesetIdentifier: number, playerIdentifiers: string[]): Observable<any> {
    return this.httpClient.post(
      this.uriRoot + "game/create-new-game",
      {
        "GameName": newGameName,
        "RulesetIdentifier": rulesetIdentifier,
        "PlayerIdentifiers": playerIdentifiers
      })
  }

  gameAsSeenByPlayer(gameIdentifier: string, playerIdentifier: string): Observable<any> {
    return this.httpClient.get(
      this.uriRoot + "game/game-as-seen-by-player/" + encodeURIComponent(gameIdentifier) + "/" + encodeURIComponent(playerIdentifier))
  }

  sendChatMessage(gameIdentifier: string, playerIdentifier: string, chatText: string): Observable<any> {
    return this.httpClient.post(
      this.uriRoot + "game/player-action",
      {
        "PlayerIdentifier": playerIdentifier,
        "GameIdentifier": gameIdentifier,
        "ActionType": "chat",
        "ChatMessage": chatText
      })
  }
}
